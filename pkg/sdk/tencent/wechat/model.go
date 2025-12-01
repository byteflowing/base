package wechat

type CommonResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type LoginResp struct {
	CommonResp // 调用者不用关心这个结构
	OpenID     string
	SessionKey string
	UnionID    string
}

type GetAccessTokenResp struct {
	CommonResp         // 调用者不用关心这个结构
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type GetStableAccessTokenBody struct {
	GrantType    string `json:"grant_type"`
	AppID        string `json:"appid"`
	Secret       string `json:"secret"`
	ForceRefresh bool   `json:"force_refresh"`
}

type ResetSessionResp struct {
	CommonResp        // 调用者不用关心这个结构
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
}

type GetPhoneNumberBody struct {
	Code   string  `json:"code"`   // 微信小程序端授权的code
	OpenID *string `json:"openid"` // 可选用户openid
}

type WaterMark struct {
	Timestamp int64  `json:"timestamp"` // 用户获取手机号操作的时间戳
	AppID     string `json:"appid"`     // 小程序appid
}

type PhoneInfo struct {
	PhoneNumber     string     `json:"phoneNumber"`     // 用户绑定的手机号（国外手机号会有区号）
	PurePhoneNumber string     `json:"purePhoneNumber"` // 没有区号的手机号
	CountryCode     string     `json:"countryCode"`     // 区号
	Watermark       *WaterMark `json:"watermark"`       // 数据水印
}

type GetPhoneNumberResp struct {
	CommonResp
	PhoneInfo *PhoneInfo `json:"phone_info"`
}
