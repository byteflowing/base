package user

import (
	"context"

	"github.com/byteflowing/base/dal/model"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/go-common/redis"
)

type Authenticator interface {
	AuthType() enumsv1.AuthType
	Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error)
}

type User interface {
	SignIn(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error)
	SignOut(ctx context.Context, accessSessionID string) (err error)
	SignOutBySessionId(ctx context.Context, sessionId string) (err error)
	Refresh(ctx context.Context, refreshToken string) (newToken string, err error)
	VerifyToken(ctx context.Context, token string) (err error)
	GetActiveSignInLog(ctx context.Context, uid uint64) (logs []*model.UserSignLog, err error)
	PagingGetSignInLogs(ctx context.Context, req *PagingGetSignInLogsReq) (resp *PagingGetSignInLogsResp, err error)
}

type Impl struct {
	handlers map[AuthType]Authenticator
	rdb      *redis.Redis
}

//func NewUserService(db *query.Query) User {
//	return &Impl{}
//}
//
//func (i *Impl) SignIn(ctx context.Context, req *SignInReq) (resp *SignInResp, err error) {
//	return nil, nil
//}
//
//func (i *Impl) SignOut(ctx context.Context) (err error) {
//	return nil
//}
//
//func (i *Impl) SignOutBySessionId(ctx context.Context, sessionId string) (err error) {
//	return nil
//}
//
//func (i *Impl) Refresh(ctx context.Context, refreshToken string) (newToken string, err error) {
//	return "", nil
//}
//
//func (i *Impl) VerifyToken(ctx context.Context, token string) (err error) {
//	return nil
//}
//
//func (i *Impl) GetActiveSignInLog(ctx context.Context, uid uint64) (logs []*model.UserSignLog, err error) {
//	return nil, nil
//}
//
//func (i *Impl) PagingGetSignInLogs(ctx context.Context, req *PagingGetSignInLogsReq) (resp *PagingGetSignInLogsResp, err error) {
//	return nil, err
//}
