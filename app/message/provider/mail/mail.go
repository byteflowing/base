package mail

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/byteflowing/base/app/message/internal/queue"
	"github.com/byteflowing/base/pkg/cron"
	"github.com/byteflowing/base/pkg/limiter"
	"github.com/byteflowing/base/pkg/logx"
	"github.com/byteflowing/base/pkg/mail"
	"github.com/byteflowing/base/pkg/validator"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	limiterv1 "github.com/byteflowing/proto/gen/go/limiter/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
)

const (
	accountQuotaName     = "mail_account_quota"
	addressQuotaName     = "mail_address_quota"
	cronFormat           = "@every %ds"
	mailKeyFormat        = "%d-%s"
	accountRuleKeyFormat = "%s:%d:%s"
)

type IMail interface {
	SendMail(ctx context.Context, req *msgv1.SendMailReq) (resp *msgv1.SendMailResp, err error)
	SendMailWithoutLimit(ctx context.Context, req *msgv1.SendMailReq) (resp *msgv1.SendMailResp, err error)
	Stop()
}

type Mail struct {
	cfg        *msgv1.MailConfig
	clients    map[string]*mail.SMTP
	queue      queue.Queue
	quota      *limiter.Quota
	cron       *cron.Cron
	entryIDs   map[string]cron.EntryID
	validators *validator.Chain[*msgv1.SendMailReq, *msgv1.SendMailResp]
	cancel     context.CancelFunc
}

func New(
	cfg *configv1.Config,
	quota *limiter.Quota,
	queue queue.Queue,
	cron *cron.Cron,
) IMail {
	if cfg == nil || cfg.MessageConfig == nil {
		panic("config is nil")
	}
	if cfg.MessageConfig.Sms == nil {
		return UnimplementedMailService{}
	}
	mailConfig := cfg.MessageConfig.Mail
	clients := make(map[string]*mail.SMTP, len(mailConfig.Providers))
	for _, provider := range mailConfig.Providers {
		key := getMailKey(provider.Vendor, provider.Account)
		m, err := mail.NewSMTP(provider.Smtp)
		if err != nil {
			panic(err)
		}
		clients[key] = m
	}
	m := &Mail{
		cfg:     mailConfig,
		clients: clients,
		queue:   queue,
		quota:   quota,
		cron:    cron,
	}
	m.validators = validator.NewChain[*msgv1.SendMailReq, *msgv1.SendMailResp](m.isAllowed)
	m.validators.Add(accountQuotaName, validator.ValidateFunc[*msgv1.SendMailReq, *msgv1.SendMailResp](m.checkAccountQuota))
	m.validators.Add(addressQuotaName, validator.ValidateFunc[*msgv1.SendMailReq, *msgv1.SendMailResp](m.checkAddressQuota))
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.queue.Register(ctx, m.cfg.Topic, m.consumeHandler)
	return m
}

func (m *Mail) SendMail(ctx context.Context, req *msgv1.SendMailReq) (resp *msgv1.SendMailResp, err error) {
	resp, err = m.validators.Check(ctx, req)
	if err != nil || resp != nil {
		return resp, err
	}
	if err = m.queue.Publish(ctx, m.cfg.Topic, req); err != nil {
		return nil, err
	}
	return
}

// SendMailWithoutLimit 不对地址进行限额（供应商的限制会校验）
func (m *Mail) SendMailWithoutLimit(ctx context.Context, req *msgv1.SendMailReq) (resp *msgv1.SendMailResp, err error) {
	resp, err = m.validators.CheckWithNames(ctx, req, accountQuotaName)
	if err != nil || resp != nil {
		return resp, err
	}
	if err = m.queue.Publish(ctx, m.cfg.Topic, req); err != nil {
		return nil, err
	}
	return
}

// Stop 移除cron，断开客户端连接
func (m *Mail) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	for _, entryID := range m.entryIDs {
		m.cron.Remove(entryID)
	}
	for _, client := range m.clients {
		_ = client.Close()
	}
}

func (m *Mail) consumeHandler(ctx context.Context, msg []byte) error {
	message := &msgv1.SendMailReq{}
	if err := proto.Unmarshal(msg, message); err != nil {
		return err
	}
	smtp, err := m.getSMTP(message.Vendor, message.Account)
	if err != nil {
		return err
	}
	if !smtp.IsConnected() {
		if err = smtp.Dial(ctx); err != nil {
			return err
		}
	}
	return smtp.Send(ctx, message)
}

func (m *Mail) getSMTP(vendor enumsv1.MailVendor, account string) (*mail.SMTP, error) {
	key := getMailKey(vendor, account)
	smtp, ok := m.clients[key]
	if !ok {
		return nil, errors.New("account not exist")
	}
	return smtp, nil
}

// 邮件供应商限流的rule_key
func (m *Mail) getAccountRuleKey(req *msgv1.SendMailReq) string {
	// ${QUOTA_PREFIX}:${VENDOR}:${ACCOUNT}
	return fmt.Sprintf(accountRuleKeyFormat, m.cfg.QuotaPrefix, req.Vendor, req.Account)
}

func (m *Mail) getAddressKey() string {
	return m.cfg.QuotaPrefix
}

func (m *Mail) isAllowed(result *msgv1.SendMailResp) bool {
	if result == nil {
		return true
	}
	return result.Rule.Allowed
}

func (m *Mail) checkAccountQuota(ctx context.Context, req *msgv1.SendMailReq) (*msgv1.SendMailResp, error) {
	rule, err := m.quota.Take(ctx, &limiterv1.RuleParam{
		Target:  req.Account,
		RuleKey: m.getAccountRuleKey(req),
	})
	if err != nil {
		return nil, err
	}
	if rule != nil && !rule.Allowed {
		resp := &msgv1.SendMailResp{
			Rule: rule,
		}
		return resp, nil
	}
	return nil, nil
}

func (m *Mail) checkAddressQuota(ctx context.Context, req *msgv1.SendMailReq) (*msgv1.SendMailResp, error) {
	var addresses []*msgv1.MailAddress
	addresses = append(addresses, req.To...)
	addresses = append(addresses, req.Bcc...)
	addresses = append(addresses, req.Cc...)
	for _, addr := range addresses {
		rule, err := m.quota.Take(ctx, &limiterv1.RuleParam{
			Target:  addr.Address,
			RuleKey: m.getAddressKey(),
		})
		if err != nil {
			return nil, err
		}
		if rule != nil && !rule.Allowed {
			resp := &msgv1.SendMailResp{
				Rule: rule,
			}
			return resp, nil
		}
	}
	return nil, nil
}

func (m *Mail) cronMgr(ctx context.Context) error {
	m.entryIDs = make(map[string]cron.EntryID, len(m.cfg.Providers))
	for _, provider := range m.cfg.Providers {
		vendor := provider.Vendor
		account := provider.Account
		client, err := m.getSMTP(vendor, account)
		if err != nil {
			return err
		}
		entryID, err := m.cron.AddFunc(fmt.Sprintf(cronFormat, m.cfg.CheckConnectionInterval.Seconds), func() {
			if err := client.CheckConnection(ctx); err != nil {
				logx.Error("[message:mail] check connection error",
					zap.String("vendor", vendor.String()),
					zap.String("account", account),
					zap.Error(err),
				)
			}
		})
		if err != nil {
			return err
		}
		m.entryIDs[getMailKey(vendor, account)] = entryID
	}
	return nil
}

func getMailKey(vendor enumsv1.MailVendor, account string) string {
	return fmt.Sprintf(mailKeyFormat, vendor, account)
}
