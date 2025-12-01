package tencent

import (
	"context"

	"github.com/byteflowing/base/app/user/dal/query"
	"github.com/byteflowing/base/pkg/sdk/tencent/wechat"
	userv1 "github.com/byteflowing/proto/gen/go/user/v1"
	wechatv1 "github.com/byteflowing/proto/gen/go/wechat/v1"
)

type WechatManager struct {
	manager *wechat.Manager
}

func NewWechatManager(cfg *wechatv1.WechatConfig) *WechatManager {
	m := wechat.NewManager(cfg)
	return &WechatManager{
		manager: m,
	}
}

func (m *WechatManager) Authenticate(ctx context.Context, req *userv1.SignInReq, tx *query.Query) (*userv1.SignInResult, error) {
	panic("implement me")
	//if req == nil || req.SignInType != enumsv1.SignInType_SIGN_IN_TYPE_WECHAT_MINI {
	//	return nil, errors.New("invalid params")
	//}
	//param := req.GetWechat()
	//result, err := m.manager.SignIn(ctx, param)
	//if err != nil {
	//	return nil, err
	//}
	//return &userv1.SignInResult{
	//
	//}, nil
}
