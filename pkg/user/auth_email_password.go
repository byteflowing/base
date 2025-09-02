package user

import (
	"context"
	"errors"

	"github.com/byteflowing/base/dal/query"
	"gorm.io/gorm"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/go-common/crypto"
)

type EmailPassword struct {
	passHasher *crypto.PasswordHasher
	repo       Repo
	jwtService *JwtService
	limiter    Limiter
}

func NewEmailPassword(passHasher *crypto.PasswordHasher, repo Repo, query *query.Query, jwtService *JwtService, limiter Limiter) Authenticator {
	return &EmailPassword{
		passHasher: passHasher,
		repo:       repo,
		jwtService: jwtService,
		limiter:    limiter,
	}
}

func (e *EmailPassword) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_EMAIL_PASSWORD
}

func (e *EmailPassword) SignUp(ctx context.Context, tx *query.Query, req *userv1.SignUpReq) (*userv1.SignUpResp, error) {
	return nil, ecode.ErrUnImplemented
}

func (e *EmailPassword) SignIn(ctx context.Context, tx *query.Query, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != e.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	userBasic, err := e.repo.GetUserBasicByEmail(ctx, tx, req.Biz, req.Identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrEmailNotExist
		}
		return nil, err
	}
	return checkPasswordAndGenToken(ctx, tx, req, userBasic, e.jwtService, e.limiter, e.passHasher)
}

func (e *EmailPassword) SignOut(ctx context.Context, tx *query.Query, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error) {
	if req.AuthType != e.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	return signOutBySessionId(ctx, req, e.repo, tx, e.jwtService)
}
