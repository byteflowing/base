package user

import (
	"time"

	"github.com/byteflowing/base/dal/model"
)

type AuthType int

const (
	AuthTypeNamePassword AuthType = iota + 1
	AuthTypeEmailPassword
	AuthTypePhoneCaptcha
	AuthTypeEmailCaptcha
	AuthTypeWechat
)

type SessionStatus int

const (
	SessionStatusSignIn SessionStatus = iota + 1
	SessionStatusSignOut
	SessionStatusKickedOut
	SessionStatusForceSignOut
)

type Status int

const (
	StatusOk = iota + 1
	StatusDisabled
)

type Source int

const (
	SourceWebsite Source = iota + 1 // 网站
	SourceApp                       // app
	SourceWechat                    // 微信
	SourceAdmin                     // 管理员添加
)

type AuthStatus int

const (
	AuthStatusOk AuthStatus = iota + 1
	AuthStatusDisabled
)

type SignInReq struct {
	// 认证类型，例如 "password", "email", "wechat"
	AuthType AuthType
	// 用户名、邮箱、手机号、openId 等
	// 根据AuthType这里可以是邮箱，账号，验证码token，openId等
	Identifier string
	// 密码、验证码、token 等
	// 如果是验证码登录，这里是验证码
	Credential string
	// 如果通过验证码登录，这里是验证码在redis中的key
	CaptchaToken string
	// 登录 IP
	IP *string
	// 用户位置
	Location *string
	// UA 信息
	UserAgent *string
	// 设备信息
	Device *string
	// jwt除标准字段外的自定义字段
	ExtraJwtClaims interface{}
}

type SignInResp struct {
	// jwt token
	// claims为jwt的RegisteredClaims+'x_'开头的自定义字段
	// iss 签发方 "user-service"
	// sub user表number字段
	// aud 暂时不用
	// exp 过期时间秒级时间戳
	// nbf 暂时使用签发时间填充(立刻生效)秒级时间戳
	// iat 签发时间秒级时间戳
	// jti 使用uuid填充代表session_id
	// 自定义字段，调用方自行填充和使用
	AccessToken  string
	RefreshToken string
}

type PagingGetSignInLogsReq struct {
	Page     uint32
	Size     uint32
	Asc      bool
	UID      *uint64
	AuthType []AuthType
	Status   []SessionStatus
}

type PagingGetSignInLogsResp struct {
	TotalCount  uint64
	TotalPages  uint32
	PageSize    uint32
	PageNum     uint32
	CurrentPage uint32
	Logs        []*model.UserSignLog
}

type Session struct {
	ID        string
	UID       uint64
	IP        string
	Device    string
	CreatedAt uint64
	ExpiresAt uint64
}

type SessionItem struct {
	SessionID string
	TTL       time.Duration
}
