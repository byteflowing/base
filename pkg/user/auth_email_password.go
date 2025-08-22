package user

import (
	"context"

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

func NewEmailPassword(passHasher *crypto.PasswordHasher, repo Repo, jwtService *JwtService, limiter Limiter) Authenticator {
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

func (e *EmailPassword) Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != e.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	userBasic, err := e.repo.GetUserBasicByEmail(ctx, req.Identifier)
	if err != nil {
		return nil, err
	}
	return checkPasswordAndGenToken(ctx, req, userBasic, e.jwtService, e.limiter, e.passHasher)
}
