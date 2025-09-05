package common

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	commonv1 "github.com/byteflowing/base/gen/common/v1"
	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/3rd/tencent/mini"
	"github.com/byteflowing/go-common/trans"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	renewAccessTokenBefore = 240
	syncTokenPoint         = 60
)

type WechatManager struct {
	clients      map[string]*mini.Client
	accessTokens map[string]*commonv1.WechatGetAccessToken
	mux          sync.RWMutex
	sf           *singleflight.Group
}

func NewWechatManager(c *configv1.Wechat) *WechatManager {
	if c == nil {
		return nil
	}
	count := len(c.Credentials)
	clients := make(map[string]*mini.Client, count)
	accessTokens := make(map[string]*commonv1.WechatGetAccessToken, count)
	sf := new(singleflight.Group)
	for _, credential := range c.Credentials {
		clients[credential.Appid] = mini.NewMiniClient(&mini.Opts{
			AppID:  credential.Appid,
			Secret: credential.Secret,
		})
	}
	return &WechatManager{
		clients:      clients,
		accessTokens: accessTokens,
		sf:           sf,
	}
}

func (w *WechatManager) WechatSignIn(ctx context.Context, req *commonv1.WechatSignInReq) (resp *commonv1.WechatSignInResp, err error) {
	client, err := w.getClient(req.Appid)
	if err != nil {
		return nil, err
	}
	res, err := client.WechatLogin(ctx, &mini.WechatLoginReq{Code: req.Code})
	if err != nil {
		return nil, err
	}
	resp = &commonv1.WechatSignInResp{
		Appid:      req.Appid,
		Openid:     res.OpenID,
		SessionKey: res.SessionKey,
		UnionId:    res.UnionID,
	}
	return resp, nil
}

func (w *WechatManager) WechatCheckSignStatus(ctx context.Context, req *commonv1.WechatCheckSignInStatusReq) (resp *commonv1.WechatCheckSignInStatusResp, err error) {
	client, err := w.getClient(req.Appid)
	if err != nil {
		return nil, err
	}
	res, err := w.getAccessToken(ctx, req.Appid)
	if err != nil {
		return nil, err
	}
	ok, err := client.CheckLoginStatus(ctx, res.AccessToken, req.SessionKey, req.Openid)
	if err != nil {
		return nil, err
	}
	resp = &commonv1.WechatCheckSignInStatusResp{Ok: ok}
	return resp, nil
}

func (w *WechatManager) WechatGetPhoneNumber(ctx context.Context, req *commonv1.WechatGetPhoneNumberReq) (resp *commonv1.WechatGetPhoneNumberResp, err error) {
	client, err := w.getClient(req.Appid)
	if err != nil {
		return nil, err
	}
	res, err := w.getAccessToken(ctx, req.Appid)
	if err != nil {
		return nil, err
	}
	phoneRes, err := client.GetPhoneNumber(ctx, &mini.GetPhoneNumberReq{
		AccessToken: res.AccessToken,
		Code:        req.Code,
		OpenID:      trans.Ref(req.Openid),
	})
	if err != nil {
		return nil, err
	}
	if phoneRes.PhoneInfo != nil {
		resp = &commonv1.WechatGetPhoneNumberResp{
			PhoneNumber:     phoneRes.PhoneInfo.PhoneNumber,
			PurePhoneNumber: phoneRes.PhoneInfo.PurePhoneNumber,
			CountryCode:     phoneRes.PhoneInfo.CountryCode,
		}
	}
	return resp, nil
}

func (w *WechatManager) getClient(appid string) (*mini.Client, error) {
	client, ok := w.clients[appid]
	if !ok {
		return nil, errors.New("appid not found")
	}
	return client, nil
}

func (w *WechatManager) getAccessToken(ctx context.Context, appid string) (resp *commonv1.WechatGetAccessToken, err error) {
	if resp, ok := w.getAccessTokenFromMap(appid); ok {
		if w.needRefreshAccessToken(resp) {
			go func() {
				defer w.sf.Forget(appid)
				_, err, _ := w.sf.Do(appid, func() (interface{}, error) {
					return w.refreshAccessToken(ctx, appid)
				})
				if err != nil {
					log.Printf("refreshAccessToken error: %v", err)
				}
			}()
		}
		return resp, nil
	}
	v, err, _ := w.sf.Do(appid, func() (interface{}, error) {
		return w.refreshAccessToken(ctx, appid)
	})
	w.sf.Forget(appid)
	if err != nil {
		return nil, err
	}
	resp = v.(*commonv1.WechatGetAccessToken)
	return
}

func (w *WechatManager) refreshAccessToken(ctx context.Context, appid string) (resp *commonv1.WechatGetAccessToken, err error) {
	w.mux.Lock()
	defer w.mux.Unlock()
	resp, ok := w.accessTokens[appid]
	if ok && !w.needRefreshAccessToken(resp) {
		return resp, nil
	}
	res, err := w.getAccessTokenFromTencent(ctx, appid)
	if err != nil {
		return nil, err
	}
	resp = &commonv1.WechatGetAccessToken{
		AccessToken: res.AccessToken,
		Expiration:  timestamppb.New(time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)),
	}
	w.accessTokens[appid] = resp
	return resp, nil
}

func (w *WechatManager) needRefreshAccessToken(resp *commonv1.WechatGetAccessToken) bool {
	expireAt := resp.Expiration.AsTime().Unix()
	return expireAt-time.Now().Unix() <= renewAccessTokenBefore
}

func (w *WechatManager) getAccessTokenFromMap(appid string) (resp *commonv1.WechatGetAccessToken, ok bool) {
	w.mux.RLock()
	defer w.mux.RUnlock()
	resp, ok = w.accessTokens[appid]
	if !ok {
		return nil, false
	}
	// 如果即将过期为了请求不报错，这里返回false，触发同步更新
	now := time.Now().Unix()
	diff := resp.Expiration.AsTime().Unix() - now
	if diff < syncTokenPoint {
		return nil, false
	}
	return
}

func (w *WechatManager) getAccessTokenFromTencent(ctx context.Context, appid string) (resp *mini.GetAccessTokenResp, err error) {
	client, err := w.getClient(appid)
	if err != nil {
		return nil, err
	}
	return client.GetStableAccessToken(ctx, &mini.GetAccessTokenReq{})
}
