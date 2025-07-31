package user

import (
	"github.com/byteflowing/base/dal/model"
	"time"
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
	AuthType   Type              // 认证类型，例如 "password", "email", "wechat"
	Identifier string            // 用户名、邮箱、手机号、openId 等
	Credential string            // 密码、验证码、token 等
	Metadata   map[string]string // 第三方补充参数，如 unionId, appId
	IP         string            // 登录 IP
	Location   string            // 用户位置
	UserAgent  string            // UA 信息
	Device     string            // 设备信息
}

type LoginResp struct {
	Token string // jwt token
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
