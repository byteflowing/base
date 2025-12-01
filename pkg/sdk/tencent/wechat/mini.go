package wechat

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/byteflowing/base/pkg/httpx"
	"github.com/byteflowing/go-common/jsonx"
	"github.com/byteflowing/go-common/trans"
	wechatv1 "github.com/byteflowing/proto/gen/go/wechat/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	code2SessionURL         = "https://api.weixin.qq.com/sns/jscode2session"
	checkSessionKeyURl      = "https://api.weixin.qq.com/wxa/checksession"
	getAccessTokenURL       = "https://api.weixin.qq.com/cgi-bin/token"
	getStableAccessTokenURL = "https://api.weixin.qq.com/cgi-bin/stable_token"
	resetSessionKeyURL      = "https://api.weixin.qq.com/wxa/resetusersessionkey"
	getPhoneNumberURL       = "https://api.weixin.qq.com/wxa/business/getuserphonenumber"

	loginGrantType       = "authorization_code"
	accessTokenGrantType = "client_credential"

	sigMethodHMACSHA256 = "hmac_sha256"
)

type Client struct {
	AppID      string // 小程序 appId
	Secret     string // 小程序 appSecret
	httpClient *http.Client
}

func NewMiniClient(opts *wechatv1.WechatCredential) *Client {
	return &Client{
		AppID:      opts.Appid,
		Secret:     opts.Secret,
		httpClient: httpx.Default(),
	}
}

// MiniLogin 小程序登录
// 文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/code2Session.html
func (mini *Client) MiniLogin(ctx context.Context, req *wechatv1.WechatSignInReq) (resp *wechatv1.WechatSignInResp, err error) {
	params := url.Values{}
	params.Add("js_code", req.Code)
	params.Add("appid", mini.AppID)
	params.Add("secret", mini.Secret)
	params.Add("grant_type", loginGrantType)
	reqURL := fmt.Sprintf("%s?%s", code2SessionURL, params.Encode())
	body, err := mini.request(reqURL)
	if err != nil {
		return nil, err
	}
	res := &LoginResp{}
	if err = jsonx.Unmarshal(body, resp); err != nil {
		return nil, err
	}
	if err = mini.checkWechatErr(res.ErrCode, res.ErrMsg); err != nil {
		return nil, err
	}
	resp = &wechatv1.WechatSignInResp{
		Appid:      req.Appid,
		Openid:     res.OpenID,
		SessionKey: res.SessionKey,
		UnionId:    res.UnionID,
	}
	return
}

// GetAccessToken : 获取接口调用凭据
// 建议优先使用GetStableAccessToken接口
// 文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getAccessToken.html
func (mini *Client) GetAccessToken(ctx context.Context, _ *wechatv1.WechatGetAccessTokenReq) (resp *wechatv1.WechatGetAccessTokenResp, err error) {
	params := url.Values{}
	params.Add("appid", mini.AppID)
	params.Add("secret", mini.Secret)
	params.Add("grant_type", accessTokenGrantType)
	reqURL := fmt.Sprintf("%s?%s", getAccessTokenURL, params.Encode())
	body, err := mini.request(reqURL)
	if err != nil {
		return nil, err
	}
	res := &GetAccessTokenResp{}
	if err = jsonx.Unmarshal(body, res); err != nil {
		return nil, err
	}
	if err = mini.checkWechatErr(res.ErrCode, res.ErrMsg); err != nil {
		return nil, err
	}
	resp = &wechatv1.WechatGetAccessTokenResp{
		AccessToken: res.AccessToken,
		Expiration:  timestamppb.New(time.Unix(int64(res.ExpiresIn), 0)),
	}
	return
}

// GetStableAccessToken : 获取接口调用凭证
// 与GetAccessToken的区别是：1. 接口允许调用次数比较多 2.在过期前每次调用返回的token都一样(force_refresh:false) 3.小项目可以直接每次调用该接口获取token，免去部署中心服务
// 建议还是存在redis或者本地缓存中，把过期时间设置为过期前五分钟，然后重新获取
// 文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getStableAccessToken.html
func (mini *Client) GetStableAccessToken(ctx context.Context, req *wechatv1.WechatGetAccessTokenReq) (resp *wechatv1.WechatGetAccessTokenResp, err error) {
	body := &GetStableAccessTokenBody{
		GrantType:    accessTokenGrantType,
		AppID:        mini.AppID,
		Secret:       mini.Secret,
		ForceRefresh: req.ForceFresh,
	}
	bodyBytes, err := jsonx.Marshal(body)
	if err != nil {
		return nil, err
	}
	respBody, err := mini.postRequest(getStableAccessTokenURL, bodyBytes)
	if err != nil {
		return nil, err
	}
	res := &GetAccessTokenResp{}
	if err = jsonx.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	if err = mini.checkWechatErr(res.ErrCode, res.ErrMsg); err != nil {
		return nil, err
	}
	resp = &wechatv1.WechatGetAccessTokenResp{
		AccessToken: res.AccessToken,
		Expiration:  timestamppb.New(time.Unix(int64(res.ExpiresIn), 0)),
	}
	return
}

// CheckLoginStatus : 检测登录状态
// 文档: https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/checkSessionKey.html
func (mini *Client) CheckLoginStatus(ctx context.Context, accessToken, sessionKey, openID string) (ok bool, err error) {
	signature := mini.signSessionKey(sessionKey)
	params := url.Values{}
	params.Add("openid", openID)
	params.Add("access_token", accessToken)
	params.Add("signature", signature)
	params.Add("sig_method", sigMethodHMACSHA256)
	reqURL := fmt.Sprintf("%s?%s", checkSessionKeyURl, params.Encode())
	body, err := mini.request(reqURL)
	if err != nil {
		return false, err
	}
	resp := &CommonResp{}
	if err = jsonx.Unmarshal(body, resp); err != nil {
		return false, err
	}
	if resp.ErrCode != 0 {
		return false, nil
	}
	return true, nil
}

// ResetSessionKey 重置登录态
// 文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/ResetUserSessionKey.html
func (mini *Client) ResetSessionKey(ctx context.Context, req *wechatv1.WechatResetSessionKeyReq) (resp *wechatv1.WechatResetSessionKeyResp, err error) {
	signature := mini.signSessionKey(req.SessionKey)
	params := url.Values{}
	params.Add("access_token", req.AccessToken)
	params.Add("openid", req.Openid)
	params.Add("signature", signature)
	params.Add("sig_method", sigMethodHMACSHA256)
	reqURL := fmt.Sprintf("%s?%s", resetSessionKeyURL, params.Encode())
	body, err := mini.request(reqURL)
	if err != nil {
		return nil, err
	}
	res := &ResetSessionResp{}
	if err = jsonx.Unmarshal(body, resp); err != nil {
		return nil, err
	}
	if err = mini.checkWechatErr(res.ErrCode, res.ErrMsg); err != nil {
		return nil, err
	}
	resp = &wechatv1.WechatResetSessionKeyResp{
		Openid:     res.OpenID,
		SessionKey: res.SessionKey,
	}
	return
}

func (mini *Client) GetPhoneNumber(ctx context.Context, req *wechatv1.WechatGetPhoneNumberReq) (resp *wechatv1.WechatGetPhoneNumberResp, err error) {
	params := url.Values{}
	params.Add("access_token", req.AccessToken)
	_url := fmt.Sprintf("%s?%s", getPhoneNumberURL, params.Encode())
	body := &GetPhoneNumberBody{
		Code:   req.Code,
		OpenID: trans.String(req.Openid),
	}
	bodyBytes, err := jsonx.Marshal(body)
	if err != nil {
		return nil, err
	}
	respBody, err := mini.postRequest(_url, bodyBytes)
	if err != nil {
		return nil, err
	}
	res := &GetPhoneNumberResp{}
	if err = jsonx.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	if err = mini.checkWechatErr(res.ErrCode, res.ErrMsg); err != nil {
		return nil, err
	}
	resp = &wechatv1.WechatGetPhoneNumberResp{
		PhoneNumber:     res.PhoneInfo.PhoneNumber,
		PurePhoneNumber: res.PhoneInfo.PurePhoneNumber,
		CountryCode:     res.PhoneInfo.CountryCode,
	}
	return
}

func (mini *Client) signSessionKey(sessionKey string) (signature string) {
	h := hmac.New(sha256.New, []byte(sessionKey))
	h.Write([]byte(sessionKey))
	return hex.EncodeToString(h.Sum(nil))
}

func (mini *Client) checkWechatErr(errCode int, errMsg string) (err error) {
	if errCode == 0 {
		return nil
	}
	return fmt.Errorf("%d:%s", errCode, errMsg)
}

func (mini *Client) request(reqURL string) ([]byte, error) {
	response, err := mini.httpClient.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func (mini *Client) postRequest(reqUrl string, body []byte) ([]byte, error) {
	reader := bytes.NewReader(body)
	response, err := mini.httpClient.Post(reqUrl, "application/json", reader)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}
