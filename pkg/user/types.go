package user

import (
	"time"

	"github.com/byteflowing/base/dal/model"
)

type Type int

const (
	AuthTypeUserNamePassword Type = iota + 1
	AuthTypeEmailPassword
	AuthTypePhoneCaptcha
	AuthTypeWechat
)

type LoginStatus int

const (
	LoginStatusOnline LoginStatus = iota + 1
	LoginStatusOffline
)

type LoginReq struct {
	// 认证类型，例如 "password", "email", "wechat"
	AuthType Type
	// 用户名、邮箱、手机号、openId 等
	Identifier string
	// 密码、验证码、token 等
	Credential string
	// 第三方补充参数，如 unionId, appId
	Metadata map[string]string
	// 登录 IP
	IP string
	// 用户位置
	Location string
	// UA 信息
	UserAgent string
	// 设备信息
	Device string
	// jwt除标准字段外的自定义字段
	// 自定义字段必须以“x_”开头，且必须小写snake_case
	// e.g. x_user_id x_app_id x_uid
	ExtraJwtClaims map[string]any
}

type LoginResp struct {
	// jwt token
	// claims为jwt的RegisteredClaims+'x_'开头的自定义字段
	// iss 签发方 "user-service"
	// sub user表number字段
	// aud 接收方，“wechat”, "web", "app"...
	// exp 过期时间秒级时间戳
	// nbf 暂时使用签发时间填充(立刻生效)秒级时间戳
	// iat 签发时间秒级时间戳
	// jti 使用uuid填充代表session_id
	// x_ 自定义字段，调用方自行填充和使用
	Token string
}

type PagingGetLoginLogsReq struct {
	Page     uint32
	Size     uint32
	Asc      bool
	UID      *uint64
	AuthType []Type
	Status   []LoginStatus
}

type PagingGetLoginLogsResp struct {
	TotalCount  uint64
	TotalPages  uint32
	PageSize    uint32
	PageNum     uint32
	CurrentPage uint32
	Logs        []*model.UserLoginLog
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
