package message

import (
	"context"
	"github.com/byteflowing/base/biz/config"
	"github.com/byteflowing/base/biz/dal/query"
	"github.com/byteflowing/base/kitex_gen/base"
	"github.com/byteflowing/go-common/redis"
)

type Opts struct {
	Conf  *config.MessageConfig
	Db    *query.Query
	Redis *redis.Redis
}

type Impl struct {
	conf  *config.MessageConfig
	db    *query.Query
	redis *redis.Redis
}

func New(opts *Opts) Message {
	return &Impl{
		conf:  opts.Conf,
		db:    opts.Db,
		redis: opts.Redis,
	}
}

func (i Impl) SendCaptcha(ctx context.Context, req *base.SendCaptchaReq) (resp *base.SendCaptchaResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) VerifyCaptcha(ctx context.Context, req *base.VerifyCaptchaReq) (resp *base.VerifyCaptchaResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) PagingGetSmsMessages(ctx context.Context, req *base.PagingGetSmsMessagesReq) (resp *base.PagingGetSmsMessagesResp, err error) {
	//TODO implement me
	panic("implement me")
}
