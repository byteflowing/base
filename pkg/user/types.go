package user

import (
	"time"

	"github.com/byteflowing/base/dal/model"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/golang-jwt/jwt/v5"
)

type SessionItem struct {
	SessionID string
	TTL       time.Duration
}

type JwtClaims struct {
	Uid      int64       `json:"uid"`
	Type     int32       `json:"type"`
	AuthType int32       `json:"auth_type"`
	OpenId   string      `json:"open_id"`
	UnionId  string      `json:"union_id"`
	AppId    string      `json:"app_id"`
	Extra    interface{} `json:"extra"`
	jwt.RegisteredClaims
}

type GenerateJwtReq struct {
	RefreshToken   string
	UserBasic      *model.UserBasic
	SignInReq      *userv1.SignInReq
	ExtraJwtClaims interface{}
	AuthType       enumsv1.AuthType
	AppId          string
	OpenId         string
	UnionId        string
}
