package sms

import (
	"context"

	"github.com/byteflowing/base/kitex_gen/base"
	"github.com/byteflowing/go-common/jsonx"
	"github.com/cloudwego/kitex/pkg/klog"
)

func getCaptchaType(t *base.CaptchaType) *int16 {
	if t == nil {
		return nil
	}
	typ := int16(*t)
	return &typ
}

func marshalParams(ctx context.Context, params map[string]string) string {
	p, err := jsonx.MarshalToString(params)
	if err != nil {
		klog.CtxErrorf(ctx, "marshal params error:%v", err)
		return ""
	}
	return p
}
