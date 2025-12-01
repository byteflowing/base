package service

import (
	"context"
	"fmt"
	"time"

	"github.com/byteflowing/base/pkg/queue/asynqx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/byteflowing/base/app/message/provider/mail"
	"github.com/byteflowing/base/app/message/provider/sms"
	"github.com/byteflowing/base/app/message/queue"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/base/pkg/limiter"
	mailService "github.com/byteflowing/base/pkg/mail"
	"github.com/byteflowing/base/pkg/quota"
	"github.com/byteflowing/base/pkg/redis"
	"github.com/byteflowing/base/pkg/utils/idx"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
	typesv1 "github.com/byteflowing/proto/gen/go/types/v1"
	"github.com/hibiken/asynq"
)

type taskType string

const (
	taskTypeSend taskType = "send"
	taskTypeSave taskType = "save"
)

const (
	taskNameFormat = "%s:%s" // send:sender

	dailyQuotaKeyPrefixFormat   = "%s:%d:%s"       // prefix:sender:target
	rateLimiterPrefixFormat     = "%s:%d:%d:%d:%s" // prefix:sender:vendor:interface:account
	slidingQuotaKeyPrefixFormat = "%s:%d:%d"       // prefix:sender:scene
	captchaPrefixFormat         = "%s:%d"
)

const (
	dailyTTL = 24 * time.Hour
)

type rateLimiterConfig struct {
	quota    int64
	interval time.Duration
}

type MessageService struct {
	cfg                 *msgv1.MessageConfig
	queue               *queue.Queue
	captcha             map[enumv1.MessageSenderType]*captcha.MessageCaptcha
	smsProviders        map[enumv1.MessageSenderVendor]map[string]sms.ISms
	mailProviders       map[enumv1.MessageSenderVendor]map[string]mail.IMail
	slidingRules        map[enumv1.MessageSenderType]map[enumv1.MessageSceneType][]*quota.SlidingRule
	rateLimiterCapacity map[enumv1.MessageSenderType]map[enumv1.MessageSenderVendor]map[string]map[enumv1.MessageInterface]*rateLimiterConfig
	dailyQuotaLimiter   *quota.FixedQuota
	slidingQuota        *quota.SlidingQuota
	rateLimiter         *limiter.DistributedRateLimiter
	smsAccountsMapping  []*msgv1.VendorAccountMapping
	mailAccountsMapping []*msgv1.VendorAccountMapping
	msgv1.UnimplementedMessageServiceServer
}

func NewMessageService(
	cfg *configv1.Config,
	rdb *redis.Redis,
) *MessageService {
	if cfg.Message == nil || cfg.AsynqServer == nil {
		panic("message and asynq server configuration required")
	}
	s := &MessageService{
		cfg:                 cfg.Message,
		queue:               queue.NewQueue(rdb, cfg.AsynqServer),
		captcha:             make(map[enumv1.MessageSenderType]*captcha.MessageCaptcha),
		slidingRules:        make(map[enumv1.MessageSenderType]map[enumv1.MessageSceneType][]*quota.SlidingRule),
		rateLimiterCapacity: make(map[enumv1.MessageSenderType]map[enumv1.MessageSenderVendor]map[string]map[enumv1.MessageInterface]*rateLimiterConfig),
		dailyQuotaLimiter:   quota.NewFixedQuota(rdb),
		slidingQuota:        quota.NewSlidingQuota(rdb),
		rateLimiter:         limiter.NewDistributedRateLimiter(rdb),
	}
	initSms(cfg, rdb, s)
	initMail(cfg, rdb, s)
	return s
}

func (m *MessageService) SendSms(ctx context.Context, req *msgv1.SendSmsReq) (*msgv1.SendSmsResp, error) {
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS
	if err := m.checkAccountAvailable(t, req.Vendor, req.Account); err != nil {
		return nil, err
	}
	phones := []string{m.joinPhoneNumber(req.PhoneNumber)}
	result, err := m.checkDailyQuota(ctx, t, phones, int(m.cfg.Sms.DailyQuota))
	if err != nil {
		return nil, err
	}
	if result != nil {
		return &msgv1.SendSmsResp{Result: result}, nil
	}
	return m.enqueueSms(ctx, req, asynq.MaxRetry(int(m.cfg.Sms.MaxTries)))
}

func (m *MessageService) SendSmsWithoutLimit(ctx context.Context, req *msgv1.SendSmsReq) (*msgv1.SendSmsResp, error) {
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS
	if err := m.checkAccountAvailable(t, req.Vendor, req.Account); err != nil {
		return nil, err
	}
	return m.enqueueSms(ctx, req, asynq.MaxRetry(int(m.cfg.Sms.MaxTries)))
}

func (m *MessageService) QuerySmsStatistics(ctx context.Context, req *msgv1.QuerySmsStatisticsReq) (*msgv1.QuerySmsStatisticsResp, error) {
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS
	if err := m.checkAccountAvailable(t, req.Vendor, req.Account); err != nil {
		return nil, err
	}
	if err := m.wait(ctx, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS, req.Vendor, enumv1.MessageInterface_MESSAGE_INTERFACE_SMS_QUERY_SEND_STATISTICS, req.Account); err != nil {
		return nil, err
	}
	return m.getISms(req.Vendor, req.Account).QuerySmsStatistics(ctx, req)
}

func (m *MessageService) QuerySmsSendDetail(ctx context.Context, req *msgv1.QuerySmsSendDetailReq) (*msgv1.QuerySmsSendDetailResp, error) {
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS
	if err := m.checkAccountAvailable(t, req.Vendor, req.Account); err != nil {
		return nil, err
	}
	if err := m.wait(ctx, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS, req.Vendor, enumv1.MessageInterface_MESSAGE_INTERFACE_SMS_QUERY_SEND_DETAILS, req.Account); err != nil {
		return nil, err
	}
	return m.getISms(req.Vendor, req.Account).QuerySmsSendDetail(ctx, req)
}

func (m *MessageService) GetSmsAccounts(ctx context.Context, req *msgv1.GetSmsAccountsReq) (*msgv1.GetSmsAccountsResp, error) {
	return &msgv1.GetSmsAccountsResp{Mappings: m.smsAccountsMapping}, nil
}

func (m *MessageService) SendMail(ctx context.Context, req *msgv1.SendMailReq) (*msgv1.SendMailResp, error) {
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL
	if err := m.checkAccountAvailable(t, req.Vendor, req.Account); err != nil {
		return nil, err
	}
	var targets []string
	for _, addr := range req.To {
		targets = append(targets, addr.Address)
	}
	for _, addr := range req.Cc {
		targets = append(targets, addr.Address)
	}
	for _, addr := range req.Bcc {
		targets = append(targets, addr.Address)
	}
	result, err := m.checkDailyQuota(ctx, t, targets, int(m.cfg.Mail.DailyQuota))
	if err != nil {
		return nil, err
	}
	if result != nil {
		return &msgv1.SendMailResp{Result: result}, nil
	}
	return m.enqueueMail(ctx, req, asynq.MaxRetry(int(m.cfg.Mail.MaxTries)))
}

func (m *MessageService) SendMailWithoutLimit(ctx context.Context, req *msgv1.SendMailReq) (*msgv1.SendMailResp, error) {
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL
	if err := m.checkAccountAvailable(t, req.Vendor, req.Account); err != nil {
		return nil, err
	}
	return m.enqueueMail(ctx, req, asynq.MaxRetry(int(m.cfg.Mail.MaxTries)))
}

func (m *MessageService) SendCaptcha(ctx context.Context, req *msgv1.SendCaptchaReq) (*msgv1.SendCaptchaResp, error) {
	if req.MessageSenderType == enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS {
		return m.sendSmsCaptcha(ctx, req)
	}
	if req.MessageSenderType == enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL {
		return m.sendMailCaptcha(ctx, req)
	}
	return nil, ecode.ErrMsgSenderUnsupported
}

func (m *MessageService) QueryMailSendStatistics(ctx context.Context, req *msgv1.QueryMailSendStatisticsReq) (*msgv1.QueryMailSendStatisticsResp, error) {
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL
	if err := m.checkAccountAvailable(t, req.Vendor, req.Account); err != nil {
		return nil, err
	}
	if err := m.wait(ctx, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL, req.Vendor, enumv1.MessageInterface_MESSAGE_INTERFACE_MAIL_SEND_STATISTICS, req.Account); err != nil {
		return nil, err
	}
	return m.getIMail(req.Vendor, req.Account).QueryMailSendStatistics(ctx, req)
}

func (m *MessageService) QueryMailSendDetails(ctx context.Context, req *msgv1.QueryMailSendDetailsReq) (*msgv1.QueryMailSendDetailsResp, error) {
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL
	if err := m.checkAccountAvailable(t, req.Vendor, req.Account); err != nil {
		return nil, err
	}
	if err := m.wait(ctx, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL, req.Vendor, enumv1.MessageInterface_MESSAGE_INTERFACE_MAIL_SEND_DETAILS, req.Account); err != nil {
		return nil, err
	}
	return m.getIMail(req.Vendor, req.Account).QueryMailSendDetails(ctx, req)
}

func (m *MessageService) QueryMailSendTracks(ctx context.Context, req *msgv1.QueryMailTracksReq) (*msgv1.QueryMailTracksResp, error) {
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL
	if err := m.checkAccountAvailable(t, req.Vendor, req.Account); err != nil {
		return nil, err
	}
	if err := m.wait(ctx, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL, req.Vendor, enumv1.MessageInterface_MESSAGE_INTERFACE_MAIL_TRACK_LIST, req.Account); err != nil {
		return nil, err
	}
	return m.getIMail(req.Vendor, req.Account).QueryMailTracks(ctx, req)
}

func (m *MessageService) GetMailAccounts(ctx context.Context, req *msgv1.GetMailAccountsReq) (*msgv1.GetMailAccountsResp, error) {
	return &msgv1.GetMailAccountsResp{Mappings: m.mailAccountsMapping}, nil
}

func (m *MessageService) VerifyCaptcha(ctx context.Context, req *msgv1.VerifyCaptchaReq) (*msgv1.VerifyCaptchaResp, error) {
	if !m.checkTokenAvailable(req.Token) {
		return nil, ecode.ErrParams
	}
	var target string
	var q uint32
	if req.MessageSenderType == enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS {
		target = m.joinPhoneNumber(req.GetPhone())
		q = m.cfg.Captcha.SmsCaptcha.MaxTries
	} else if req.MessageSenderType == enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL {
		q = m.cfg.Captcha.MailCaptcha.MaxTries
		target = req.GetMail()
	}
	c := m.getCaptcha(req.MessageSenderType)
	res, err := c.Verify(ctx, target, req.Token, req.Captcha, req.MessageSenderType, req.SceneType)
	if err != nil {
		return nil, err
	}
	return &msgv1.VerifyCaptchaResp{Result: &msgv1.VerifyCaptchaResp_VerifyResult{
		Fails: uint32(res.Fails),
		Quota: q,
	}}, nil
}

func (m *MessageService) enqueueSms(ctx context.Context, req *msgv1.SendSmsReq, options ...asynq.Option) (*msgv1.SendSmsResp, error) {
	taskName := m.getTaskName(taskTypeSend, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS)
	if err := m.queue.EnQueue(ctx, taskName, req, options...); err != nil {
		return nil, err
	}
	return &msgv1.SendSmsResp{}, nil
}

func (m *MessageService) enqueueMail(ctx context.Context, req *msgv1.SendMailReq, options ...asynq.Option) (*msgv1.SendMailResp, error) {
	taskName := m.getTaskName(taskTypeSend, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL)
	if err := m.queue.EnQueue(ctx, taskName, req, options...); err != nil {
		return nil, err
	}
	return &msgv1.SendMailResp{}, nil
}

// asynq的回调函数
func (m *MessageService) sendSms(ctx context.Context, task *asynq.Task) error {
	req := &msgv1.SendSmsReq{}
	if err := proto.Unmarshal(task.Payload(), req); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, m.cfg.Sms.SendTimeout.AsDuration())
	defer cancel()
	if err := m.wait(ctx, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS, req.Vendor, enumv1.MessageInterface_MESSAGE_INTERFACE_SMS_SEND_SINGLE_MESSAGE, req.Account); err != nil {
		return err
	}
	return m.getISms(req.Vendor, req.Account).SendSingleSms(ctx, req)
}

// asynq的回调函数
func (m *MessageService) sendMail(ctx context.Context, task *asynq.Task) error {
	req := &msgv1.SendMailReq{}
	if err := proto.Unmarshal(task.Payload(), req); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, m.cfg.Mail.SendTimeout.AsDuration())
	defer cancel()
	if err := m.wait(ctx, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL, req.Vendor, enumv1.MessageInterface_MESSAGE_INTERFACE_MAIL_SEND_SINGLE_MESSAGE, req.Account); err != nil {
		return err
	}
	return m.getIMail(req.Vendor, req.Account).SendMail(ctx, req)
}

func (m *MessageService) sendSmsCaptcha(ctx context.Context, req *msgv1.SendCaptchaReq) (*msgv1.SendCaptchaResp, error) {
	params := req.GetSms()
	if params == nil {
		return nil, ecode.ErrParams
	}
	if err := m.checkAccountAvailable(req.MessageSenderType, params.GetVendor(), params.GetAccount()); err != nil {
		return nil, err
	}
	keyPrefix := m.getSlidingQuotaPrefix(m.cfg.Captcha.Prefix, req.MessageSenderType, req.SceneType)
	phones := []string{m.joinPhoneNumber(params.PhoneNumber)}
	rules := m.getSlidingRules(req.MessageSenderType, req.SceneType)
	if len(rules) == 0 {
		return nil, ecode.ErrMsgSceneUnsupported
	}
	result, err := m.checkSlidingQuota(ctx, keyPrefix, phones, rules)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return &msgv1.SendCaptchaResp{Result: result}, nil
	}
	result, err = m.checkDailyQuota(ctx, req.MessageSenderType, phones, int(m.cfg.Sms.DailyQuota))
	if err != nil {
		return nil, err
	}
	if result != nil {
		return &msgv1.SendCaptchaResp{Result: result}, nil
	}
	code, token, err := m.getCaptcha(req.MessageSenderType).Save(ctx, phones[0], req.MessageSenderType, req.SceneType)
	if err != nil {
		return nil, err
	}
	params.TemplateParams[req.CaptchaTemplateName] = code
	// 验证码一般都会在前端设置超时重试按钮，所以这里禁用失败重试
	deadLine := time.Now().Add(m.cfg.Captcha.SmsCaptcha.Ttl.AsDuration())
	if _, err = m.enqueueSms(ctx, params, asynq.Deadline(deadLine), asynq.MaxRetry(0), asynq.Queue(asynqx.QueueCritical)); err != nil {
		return nil, err
	}
	return &msgv1.SendCaptchaResp{Token: token}, nil
}

func (m *MessageService) sendMailCaptcha(ctx context.Context, req *msgv1.SendCaptchaReq) (*msgv1.SendCaptchaResp, error) {
	params := req.GetMail()
	if params == nil || len(params.To) != 1 {
		return nil, ecode.ErrParams
	}
	if err := m.checkAccountAvailable(req.MessageSenderType, params.GetVendor(), params.GetAccount()); err != nil {
		return nil, err
	}
	keyPrefix := m.getSlidingQuotaPrefix(m.cfg.Captcha.Prefix, req.MessageSenderType, req.SceneType)
	rules := m.getSlidingRules(req.MessageSenderType, req.SceneType)
	targets := []string{params.To[0].Address}
	if len(rules) == 0 {
		return nil, ecode.ErrMsgSceneUnsupported
	}
	result, err := m.checkSlidingQuota(ctx, keyPrefix, targets, rules)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return &msgv1.SendCaptchaResp{Result: result}, nil
	}
	result, err = m.checkDailyQuota(ctx, req.MessageSenderType, targets, int(m.cfg.Mail.DailyQuota))
	if err != nil {
		return nil, err
	}
	if result != nil {
		return &msgv1.SendCaptchaResp{Result: result}, nil
	}
	code, token, err := m.getCaptcha(req.MessageSenderType).Save(ctx, targets[0], req.MessageSenderType, req.SceneType)
	if err != nil {
		return nil, err
	}
	params.TemplateParams[req.CaptchaTemplateName] = code
	isHtml := params.ContentType == enumv1.MailContentType_MAIL_CONTENT_TYPE_HTML
	templateEngine := mailService.GoTemplateEngine{EnableHTML: isHtml}
	content, err := templateEngine.Render(params.Content, params.TemplateParams)
	if err != nil {
		return nil, err
	}
	params.Content = content
	// 验证码一般都会在前端设置超时重试按钮，所以这里禁用失败重试
	deadLine := time.Now().Add(m.cfg.Captcha.MailCaptcha.Ttl.AsDuration())
	if _, err = m.enqueueMail(ctx, params, asynq.Deadline(deadLine), asynq.MaxRetry(0), asynq.Queue(asynqx.QueueCritical)); err != nil {
		return nil, err
	}
	return &msgv1.SendCaptchaResp{Token: token}, nil
}

func (m *MessageService) wait(
	ctx context.Context,
	sender enumv1.MessageSenderType,
	vendor enumv1.MessageSenderVendor,
	iType enumv1.MessageInterface,
	account string,
) error {
	key := m.getRateLimiterKey(sender, vendor, iType, account)
	c := m.getRateLimiterCapacity(sender, vendor, iType, account)
	if err := m.rateLimiter.Wait(ctx, key, c.quota, c.interval, 1); err != nil {
		return err
	}
	return nil
}

func (m *MessageService) getISms(vendor enumv1.MessageSenderVendor, account string) sms.ISms {
	return m.smsProviders[vendor][account]
}

func (m *MessageService) getIMail(vendor enumv1.MessageSenderVendor, account string) mail.IMail {
	return m.mailProviders[vendor][account]
}

func (m *MessageService) getCaptcha(senderType enumv1.MessageSenderType) *captcha.MessageCaptcha {
	return m.captcha[senderType]
}

func (m *MessageService) getRateLimiterCapacity(sender enumv1.MessageSenderType, vendor enumv1.MessageSenderVendor, iType enumv1.MessageInterface, account string) *rateLimiterConfig {
	return m.rateLimiterCapacity[sender][vendor][account][iType]
}

func (m *MessageService) getSlidingRules(sender enumv1.MessageSenderType, scene enumv1.MessageSceneType) []*quota.SlidingRule {
	scenes, _ := m.slidingRules[sender][scene]
	return scenes
}

func (m *MessageService) getDailyQuotaKey(sender enumv1.MessageSenderType, target string) string {
	var prefix string
	switch sender {
	case enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS:
		prefix = m.cfg.Sms.DailyQuotaPrefix
	case enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL:
		prefix = m.cfg.Mail.DailyQuotaPrefix
	}
	return fmt.Sprintf(dailyQuotaKeyPrefixFormat, prefix, sender, target)
}

func (m *MessageService) getRateLimiterKey(
	sender enumv1.MessageSenderType,
	vendor enumv1.MessageSenderVendor,
	iType enumv1.MessageInterface,
	account string) string {
	var prefix string
	switch sender {
	case enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS:
		prefix = m.cfg.Sms.RateLimiterPrefix
	case enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL:
		prefix = m.cfg.Mail.RateLimiterPrefix
	}
	return fmt.Sprintf(rateLimiterPrefixFormat, prefix, sender, vendor, iType, account)
}

func (m *MessageService) getSlidingQuotaPrefix(prefix string, sender enumv1.MessageSenderType, scene enumv1.MessageSceneType) string {
	return fmt.Sprintf(slidingQuotaKeyPrefixFormat, prefix, sender, scene)
}

func (m *MessageService) getTaskName(typ taskType, sender enumv1.MessageSenderType) string {
	return fmt.Sprintf(taskNameFormat, typ, sender)
}

func (m *MessageService) checkAccountAvailable(senderType enumv1.MessageSenderType, vendor enumv1.MessageSenderVendor, account string) error {
	s, ok := m.rateLimiterCapacity[senderType]
	if !ok {
		return ecode.ErrMsgSenderUnsupported
	}
	v, ok := s[vendor]
	if !ok {
		return ecode.ErrMsgVendorUnsupported
	}
	_, ok = v[account]
	if !ok {
		return ecode.ErrMsgAccountUnsupported
	}
	return nil
}

func (m *MessageService) checkDailyQuota(
	ctx context.Context,
	sender enumv1.MessageSenderType,
	targets []string,
	q int,
) (result *typesv1.QuotaResult, err error) {
	for _, target := range targets {
		key := m.getDailyQuotaKey(sender, target)
		res, err := m.dailyQuotaLimiter.Allow(ctx, key, dailyTTL, q)
		if err != nil {
			return nil, err
		}
		if !res.Allowed {
			return &typesv1.QuotaResult{
				Quota:      int32(q),
				Current:    int32(res.Current),
				Window:     durationpb.New(time.Hour * 24),
				RetryAfter: durationpb.New(res.RetryAfter),
			}, nil
		}
	}
	return nil, nil
}

func (m *MessageService) checkSlidingQuota(ctx context.Context, prefix string, targets []string, rules []*quota.SlidingRule) (result *typesv1.QuotaResult, err error) {
	for _, target := range targets {
		res, err := m.slidingQuota.Allow(ctx, prefix, target, rules)
		if err != nil {
			return nil, err
		}
		if !res.Allowed {
			return &typesv1.QuotaResult{
				Quota:      int32(res.Rule.Quota),
				Current:    res.Current,
				Window:     durationpb.New(res.Rule.Window),
				RetryAfter: durationpb.New(res.RetryAfter),
			}, nil
		}
	}
	return nil, nil
}

func (m *MessageService) checkTokenAvailable(token string) bool {
	if err := idx.ValidateUUID(token); err != nil {
		return false
	}
	return true
}

func (m *MessageService) joinPhoneNumber(phone *typesv1.PhoneNumber) string {
	return phone.GetCountryCode() + phone.GetNumber()
}
