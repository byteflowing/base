package account

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/byteflowing/base/pkg/httpx"
	huaweiv1 "github.com/byteflowing/proto/gen/go/huawei/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	signInURL            = "https://account-api.cloud.huawei.com/oauth2/v6/quickLogin/getPhoneNumber"
	getAvatarNickNameURL = "https://account.cloud.huawei.com/rest.php?nsp_svc=GOpen.User.getInfo"
	getUserPhoneURL      = "https://account.cloud.huawei.com/rest.php?nsp_svc=GOpen.User.getInfo"
	getUserRiskLevelURL  = "https://account.cloud.huawei.com/user/getuserrisklevel"
	getUserTokenURL      = "https://oauth-login.cloud.huawei.com/oauth2/v3/token"
	refreshTokenURL      = "https://oauth-login.cloud.huawei.com/oauth2/v3/token"
	parseUserTokenURL    = "https://oauth-api.cloud.huawei.com/rest.php?nsp_fmt=JSON&nsp_svc=huawei.oauth2.user.getTokenInfo"
	cancelUserTokenURL   = "https://oauth-login.cloud.huawei.com/oauth2/v3/revoke"
	getAppTokenURL       = "https://oauth-login.cloud.huawei.com/oauth2/v3/token"
)

type Account struct {
	clientID     string // 在创建应用后，由华为开发者联盟为应用分配的唯一标识
	clientSecret string // 在创建应用后，由华为开发者联盟为应用分配的密钥（Client Secret）
	httpClient   *http.Client
}

func New(clientID, clientSecret string) *Account {
	return &Account{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   httpx.NewClient(httpx.GetDefaultConfig()),
	}
}

// SignIn 华为账号一键登录
// 文档：https://developer.huawei.com/consumer/cn/doc/harmonyos-references/account-api-get-user-info-quicklogin-by-code
func (a *Account) SignIn(ctx context.Context, req *huaweiv1.HuaweiSignInReq) (resp *huaweiv1.HuaweiSignInResp, err error) {
	body, err := protojson.Marshal(req)
	if err != nil {
		return nil, err
	}
	respBody, err := a.postRequest(signInURL, "application/json;charset=UTF-8", body)
	if err != nil {
		return nil, err
	}
	resp = &huaweiv1.HuaweiSignInResp{}
	if err = protojson.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUserAvatarNickName 获取用户头像和昵称
// 文档：https://developer.huawei.com/consumer/cn/doc/harmonyos-references/account-api-get-user-info-get-nickname-and-avatar
func (a *Account) GetUserAvatarNickName(ctx context.Context, req *huaweiv1.HuaweiGetUserAvatarNickNameReq) (resp *huaweiv1.HuaweiGetUserAvatarNickNameResp, err error) {
	data := url.Values{}
	data.Add("access_token", req.AccessToken)
	data.Add("getNickName", strconv.FormatInt(int64(req.GetNickName), 10))
	respBody, err := a.postFormRequest(getAvatarNickNameURL, data)
	if err != nil {
		return nil, err
	}
	resp = &huaweiv1.HuaweiGetUserAvatarNickNameResp{}
	if err = protojson.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUserPhone 获取用户电话
// 文档：https://developer.huawei.com/consumer/cn/doc/harmonyos-references/account-api-get-user-info-get-phone
func (a *Account) GetUserPhone(ctx context.Context, req *huaweiv1.HuaweiGetUserPhoneReq) (resp *huaweiv1.HuaweiGetUserPhoneResp, err error) {
	data := url.Values{}
	data.Add("access_token", req.AccessToken)
	respBody, err := a.postFormRequest(getUserPhoneURL, data)
	if err != nil {
		return nil, err
	}
	resp = &huaweiv1.HuaweiGetUserPhoneResp{}
	if err = protojson.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUserRiskLevel 获取用户风险等级
// 文档：https://developer.huawei.com/consumer/cn/doc/harmonyos-references/account-api-getuserrisklevel
func (a *Account) GetUserRiskLevel(ctx context.Context, req *huaweiv1.HuaweiGetUserRiskLevelReq) (resp *huaweiv1.HuaweiGetUserRiskLevelResp, err error) {
	body, err := protojson.Marshal(req)
	if err != nil {
		return nil, err
	}
	respBody, err := a.postRequest(getUserRiskLevelURL, "application/json;charset=UTF-8", body)
	if err != nil {
		return nil, err
	}
	resp = &huaweiv1.HuaweiGetUserRiskLevelResp{}
	if err = protojson.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUserToken 获取用户级token
// 文档：https://developer.huawei.com/consumer/cn/doc/harmonyos-references/account-api-obtain-user-token
func (a *Account) GetUserToken(ctx context.Context, req *huaweiv1.HuaweiGetUserTokenReq) (resp *huaweiv1.HuaweiGetUserTokenResp, err error) {
	data := url.Values{}
	data.Add("grant_type", "authorization_code")
	data.Add("client_id", a.clientID)
	data.Add("client_secret", a.clientSecret)
	data.Add("code", req.Code)
	data.Add("supportAlg", "PS256")
	respBody, err := a.postFormRequest(getUserTokenURL, data)
	if err != nil {
		return nil, err
	}
	resp = &huaweiv1.HuaweiGetUserTokenResp{}
	if err = protojson.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// RefreshUserToken 刷新用户级token
// 文档：https://developer.huawei.com/consumer/cn/doc/harmonyos-references/account-api-obtain-refresh-token
func (a *Account) RefreshUserToken(ctx context.Context, req *huaweiv1.HuaweiRefreshUserTokenReq) (resp *huaweiv1.HuaweiRefreshUserTokenResp, err error) {
	data := url.Values{}
	data.Add("grant_type", "refresh_token")
	data.Add("client_id", a.clientID)
	data.Add("client_secret", a.clientSecret)
	data.Add("refresh_token", req.RefreshToken)
	if req.Scope != "" {
		data.Add("scope", req.Scope)
	}
	data.Add("supportAlg", "PS256")
	respBody, err := a.postFormRequest(refreshTokenURL, data)
	if err != nil {
		return nil, err
	}
	resp = &huaweiv1.HuaweiRefreshUserTokenResp{}
	if err = protojson.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ParseUserAccessToken 解析用户级token
// 文档： https://developer.huawei.com/consumer/cn/doc/harmonyos-references/account-api-get-token-info
func (a *Account) ParseUserAccessToken(ctx context.Context, req *huaweiv1.HuaweiParseUserAccessTokenReq) (resp *huaweiv1.HuaweiParseUserAccessTokenResp, err error) {
	data := url.Values{}
	data.Add("access_token", req.AccessToken)
	data.Add("open_id", req.OpenId)
	respBody, err := a.postFormRequest(parseUserTokenURL, data)
	if err != nil {
		return nil, err
	}
	resp = &huaweiv1.HuaweiParseUserAccessTokenResp{}
	if err = protojson.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CancelUserToken 取消用户级凭证
// 文档：https://developer.huawei.com/consumer/cn/doc/harmonyos-references/account-api-obtain-revoke-token
func (a *Account) CancelUserToken(ctx context.Context, req *huaweiv1.HuaweiCancelUserTokenReq) (resp *huaweiv1.HuaweiCancelUserTokenResp, err error) {
	data := url.Values{}
	data.Add("token", req.Token)
	respBody, err := a.postFormRequest(cancelUserTokenURL, data)
	if err != nil {
		return nil, err
	}
	resp = &huaweiv1.HuaweiCancelUserTokenResp{}
	if err = protojson.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetAppToken 获取应用级token
// 文档：https://developer.huawei.com/consumer/cn/doc/harmonyos-references/account-api-obtain-app-token
func (a *Account) GetAppToken(ctx context.Context, req *huaweiv1.HuaweiGetAppTokenReq) (resp *huaweiv1.HuaweiGetAppTokenResp, err error) {
	data := url.Values{}
	data.Add("grant_type", "client_credential")
	data.Add("client_id", a.clientID)
	data.Add("client_secret", a.clientSecret)
	respBody, err := a.postFormRequest(getAppTokenURL, data)
	if err != nil {
		return nil, err
	}
	resp = &huaweiv1.HuaweiGetAppTokenResp{}
	if err = protojson.Unmarshal(respBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *Account) postFormRequest(url string, data url.Values) ([]byte, error) {
	response, err := a.httpClient.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func (a *Account) postRequest(reqUrl string, contentType string, body []byte) ([]byte, error) {
	reader := bytes.NewReader(body)
	response, err := a.httpClient.Post(reqUrl, contentType, reader)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}
