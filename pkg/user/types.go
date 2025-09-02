package user

import (
	"time"

	"github.com/byteflowing/base/dal/model"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/types/known/anypb"
)

type SessionItem struct {
	SessionID string
	TTL       time.Duration
}

type JwtClaims struct {
	Uid       int64      `json:"uid"`
	Number    string     `json:"number"`
	Biz       string     `json:"biz"`
	UserType  int32      `json:"user_type"`
	UserLevel int32      `json:"user_level"`
	TokenType int32      `json:"token_type"`
	AuthType  int32      `json:"auth_type"`
	OpenId    string     `json:"open_id"`
	UnionId   string     `json:"union_id"`
	AppId     string     `json:"app_id"`
	Extra     *anypb.Any `json:"extra"`
	jwt.RegisteredClaims
}

type GenerateJwtReq struct {
	UserBasic      *model.UserBasic
	SignInReq      *userv1.SignInReq
	ExtraJwtClaims *anypb.Any
	AuthType       enumsv1.AuthType
	AppId          string
	OpenId         string
	UnionId        string
}
