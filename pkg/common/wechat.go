package common

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/byteflowing/go-common/3rd/tencent/mini"
	wechatv1 "github.com/byteflowing/proto/gen/go/wechat/v1"
	"golang.org/x/sync/singleflight"
)

const (
	renewAccessTokenBefore = 240
	syncTokenPoint         = 60
)

type WechatManager struct {
	clients      map[string]*mini.Client
	accessTokens map[string]*wechatv1.WechatGetAccessTokenResp
	mux          sync.RWMutex
	sf           *singleflight.Group
}

func NewWechatManager(c *wechatv1.WechatConfig) *WechatManager {
	if c == nil {
		return nil
	}
	count := len(c.Credentials)
	clients := make(map[string]*mini.Client, count)
	accessTokens := make(map[string]*wechatv1.WechatGetAccessTokenResp, count)
	sf := new(singleflight.Group)
	for _, credential := range c.Credentials {
		clients[credential.Appid] = mini.NewMiniClient(credential)
	}
	return &WechatManager{
		clients:      clients,
		accessTokens: accessTokens,
		sf:           sf,
	}
}

func (w *WechatManager) WechatSignIn(ctx context.Context, req *wechatv1.WechatSignInReq) (resp *wechatv1.WechatSignInResp, err error) {
	client, err := w.getClient(req.Appid)
	if err != nil {
		return nil, err
	}
	return client.WechatLogin(ctx, req)
}

func (w *WechatManager) WechatCheckSignStatus(ctx context.Context, req *wechatv1.WechatCheckSignInStatusReq) (resp *wechatv1.WechatCheckSignInStatusResp, err error) {
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
	resp = &wechatv1.WechatCheckSignInStatusResp{Ok: ok}
	return resp, nil
}

func (w *WechatManager) WechatGetPhoneNumber(ctx context.Context, req *wechatv1.WechatGetPhoneNumberReq) (resp *wechatv1.WechatGetPhoneNumberResp, err error) {
	client, err := w.getClient(req.Appid)
	if err != nil {
		return nil, err
	}
	res, err := w.getAccessToken(ctx, req.Appid)
	if err != nil {
		return nil, err
	}
	return client.GetPhoneNumber(ctx, &wechatv1.WechatGetPhoneNumberReq{
		AccessToken: res.AccessToken,
		Code:        req.Code,
		Openid:      req.Openid,
		Appid:       req.Appid,
	})
}

func (w *WechatManager) getClient(appid string) (*mini.Client, error) {
	client, ok := w.clients[appid]
	if !ok {
		return nil, errors.New("appid not found")
	}
	return client, nil
}

func (w *WechatManager) getAccessToken(ctx context.Context, appid string) (resp *wechatv1.WechatGetAccessTokenResp, err error) {
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
	resp = v.(*wechatv1.WechatGetAccessTokenResp)
	return
}

func (w *WechatManager) refreshAccessToken(ctx context.Context, appid string) (resp *wechatv1.WechatGetAccessTokenResp, err error) {
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
	w.accessTokens[appid] = res
	return resp, nil
}

func (w *WechatManager) needRefreshAccessToken(resp *wechatv1.WechatGetAccessTokenResp) bool {
	expireAt := resp.Expiration.AsTime().Unix()
	return expireAt-time.Now().Unix() <= renewAccessTokenBefore
}

func (w *WechatManager) getAccessTokenFromMap(appid string) (resp *wechatv1.WechatGetAccessTokenResp, ok bool) {
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

func (w *WechatManager) getAccessTokenFromTencent(ctx context.Context, appid string) (resp *wechatv1.WechatGetAccessTokenResp, err error) {
	client, err := w.getClient(appid)
	if err != nil {
		return nil, err
	}
	return client.GetStableAccessToken(ctx, &wechatv1.WechatGetAccessTokenReq{})
}
