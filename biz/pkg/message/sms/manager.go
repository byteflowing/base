package sms

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/byteflowing/base/biz/config"
	"github.com/byteflowing/base/biz/constant"
	"github.com/byteflowing/base/biz/dal/query"
	"github.com/byteflowing/base/kitex_gen/base"
	"github.com/byteflowing/go-common/redis"
	"github.com/byteflowing/go-common/service"
	"github.com/cloudwego/kitex/pkg/klog"
)

const scanLimit = 500

type Manager struct {
	senders   map[base.SmsProvider]Sender
	store     Store
	smsConf   *config.SmsConfig
	cancel    context.CancelFunc
	queryLock map[base.SmsProvider]*sync.Mutex
}

type MangerOpts struct {
	SmsConfig *config.SmsConfig
	RDB       *redis.Redis
	DB        *query.Query
}

func NewSmsManager(opts *MangerOpts) *Manager {
	var store Store
	if opts.SmsConfig.SaveMessage {
		if opts.DB == nil {
			panic("DbStore must with db")
		}
		store = NewDbStore(opts.DB)
	} else {
		store = NewEmptyStore()
	}
	senders := make(map[base.SmsProvider]Sender, len(opts.SmsConfig.Providers))
	locks := make(map[base.SmsProvider]*sync.Mutex, len(opts.SmsConfig.Providers))
	for _, v := range opts.SmsConfig.Providers {
		senders[v.GetProvider()] = newSender(v, store)
		locks[v.GetProvider()] = &sync.Mutex{}
	}
	return &Manager{
		senders:   senders,
		store:     store,
		smsConf:   opts.SmsConfig,
		queryLock: locks,
	}
}

func (s *Manager) Start() {
	if !s.smsConf.SaveMessage {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	rg := service.NewRoutineGroup()
	for _, sender := range s.senders {
		localSender := sender
		rg.Run(func() {
			ticker := time.NewTicker(time.Duration(s.smsConf.QueryDetailInterval) * time.Second)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					lock := s.queryLock[localSender.GetProvider()]
					if lock.TryLock() {
						s.querySendDetail(ctx, localSender)
						lock.Unlock()
					}
				}
			}
		})
	}
	rg.Wait()
}

func (s *Manager) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Manager) GetSender(provider base.SmsProvider) (sender Sender, err error) {
	var ok bool
	sender, ok = s.senders[provider]
	if !ok {
		return nil, constant.ErrNotImplemented
	}
	return sender, nil
}

func newSender(conf *config.SmsProvider, store Store) Sender {
	provider := conf.GetProvider()
	switch provider {
	case base.SmsProvider_SMS_PROVIDER_ALIYUN:
		return NewAliSmsSender(conf, store)
	default:
		panic(fmt.Errorf("unknown provider %s", conf.Provider))
	}
}

func (s *Manager) querySendDetail(ctx context.Context, sender Sender) {
	errCount := 0
	maxTries := 10
	sleepTime := time.Second
	provider := sender.GetProvider()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		messages, err := s.store.GetSendingMessages(ctx, provider, scanLimit)
		if err != nil {
			klog.Errorf("query Sending Messages from db err:%v", err)
			if errCount >= maxTries {
				return
			}
			time.Sleep(sleepTime)
			errCount++
			continue
		}
		if len(messages) == 0 {
			return
		}
		for _, message := range messages {
			if err = sender.Wait(ctx, ApiQuerySendDetail); err != nil {
				// 这里出现错误很有可能是调用了ctx.cancel()
				klog.Errorf("wait for query detail token error :%v", err)
				return
			}
			if err = sender.QuerySendDetail(ctx, message); err != nil {
				klog.Errorf("query send detail error :%v", err)
			}
		}
		if len(messages) < scanLimit {
			return
		}
	}
}
