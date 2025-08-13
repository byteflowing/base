package mail

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/byteflowing/base/pkg/message/captcha"
	"github.com/byteflowing/go-common/mail"
	"github.com/byteflowing/go-common/ratelimit"
	"github.com/byteflowing/go-common/redis"
)

const (
	reconnects = 3
	retryWait  = 100 * time.Millisecond
)

type Mail interface {
	SendCaptcha(ctx context.Context, req *SendCaptchaReq) (token string, limit *captcha.LimitRule, err error)
	VerifyCaptcha(ctx context.Context, req *VerifyCaptchaReq) (ok bool, err error)
}

type Impl struct {
	clients map[Vendor]chan *mail.SMTP
	configs map[Vendor]*Provider
	limits  map[Vendor]*ratelimit.Limiter
	conns   map[Vendor]uint64
	mux     map[Vendor]*sync.RWMutex
	captcha captcha.Captcha
}

func New(rdb *redis.Redis, c *Config) *Impl {
	counts := len(c.Providers)
	clients := make(map[Vendor]chan *mail.SMTP, counts)
	configs := make(map[Vendor]*Provider, counts)
	limits := make(map[Vendor]*ratelimit.Limiter, counts)
	conns := make(map[Vendor]uint64, counts)
	mux := make(map[Vendor]*sync.RWMutex, counts)
	_captcha := captcha.New(rdb, c.Captcha)
	for _, provider := range c.Providers {
		vendor := provider.GetVendor()
		clients[vendor] = make(chan *mail.SMTP, provider.MaxConnections)
		configs[vendor] = provider
		limits[vendor] = ratelimit.NewLimiter(time.Duration(provider.LimitDuration)*time.Second, provider.LimitMax, provider.LimitMax)
		mux[vendor] = &sync.RWMutex{}
	}
	return &Impl{
		clients: clients,
		configs: configs,
		limits:  limits,
		conns:   conns,
		mux:     mux,
		captcha: _captcha,
	}
}

func (i *Impl) SendCaptcha(ctx context.Context, req *SendCaptchaReq) (token string, limit *captcha.LimitRule, err error) {
	return i.captcha.Save(ctx, req.To.Addr, req.Captcha, func() error {
		m := &mail.Mail{
			From:        req.From,
			To:          []*mail.Address{req.To},
			Subject:     req.Subject,
			ContentType: req.ContentType,
			Content:     req.Content,
		}
		return i.sendMail(ctx, req.Vendor, m)
	})
}

func (i *Impl) VerifyCaptcha(ctx context.Context, req *VerifyCaptchaReq) (ok bool, err error) {
	return i.captcha.Verify(ctx, req.Token, req.Captcha)
}

func (i *Impl) getSMTP(ctx context.Context, vendor Vendor) (*mail.SMTP, error) {
	if _, ok := i.clients[vendor]; !ok {
		return nil, fmt.Errorf("no such vendor client: %s", vendor)
	}
	select {
	case client := <-i.clients[vendor]:
		return client, nil
	default:
	}
	i.mux[vendor].Lock()
	// 达到最大连接，等待空闲连接或者超时取消
	if i.conns[vendor] >= i.configs[vendor].MaxConnections {
		i.mux[vendor].Unlock()
		select {
		case client := <-i.clients[vendor]:
			return client, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	// 先将连接计数自增解开锁，让其他协程获取
	i.conns[vendor]++
	i.mux[vendor].Unlock()
	client, err := mail.NewSMTP(i.configs[vendor].SMTP)
	if err != nil {
		i.mux[vendor].Lock()
		i.conns[vendor]--
		i.mux[vendor].Unlock()
	}
	return client, nil
}

func (i *Impl) releaseSMTP(ctx context.Context, vendor Vendor, client *mail.SMTP) {
	select {
	case i.clients[vendor] <- client:
	default:
		_ = client.Close(ctx)
		i.mux[vendor].Lock()
		i.conns[vendor]--
		i.mux[vendor].Unlock()
	}
}

func (i *Impl) sendMail(ctx context.Context, vendor Vendor, mails ...*mail.Mail) (err error) {
	if err := i.limits[vendor].Wait(ctx); err != nil {
		return err
	}
	client, err := i.getSMTP(ctx, vendor)
	if err != nil {
		return err
	}
	defer i.releaseSMTP(ctx, vendor, client)
	if !client.IsConnected() {
		if err := client.Dial(ctx); err != nil {
			return err
		}
	}
	err = client.Send(ctx, mails...)
	if err == nil {
		return nil
	}
	// 如果err不为nil为防止是连接断开这里重试几次，如果仍然失败再返回
	// 先关闭连接
	senderErrors := err
	if err = client.Close(ctx); err != nil {
		return err
	}
	for i := 0; i < reconnects; i++ {
		if err = client.Dial(ctx); err != nil {
			senderErrors = errors.Join(senderErrors, err)
			select {
			case <-time.After(retryWait):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		if err := client.Send(ctx, mails...); err != nil {
			senderErrors = errors.Join(senderErrors, err)
			select {
			case <-time.After(retryWait):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		return nil
	}
	return senderErrors
}
