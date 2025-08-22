package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/go-common/crypto"
)

type NumberPassword struct {
	passHasher *crypto.PasswordHasher
	repo       Repo
	jwtService *JwtService
	limiter    Limiter
}

func NewNumberPassword(passHasher *crypto.PasswordHasher, repo Repo, jwtService *JwtService, limiter Limiter) Authenticator {
	return &NumberPassword{
		passHasher: passHasher,
		repo:       repo,
		jwtService: jwtService,
		limiter:    limiter,
	}
}

func (n *NumberPassword) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_NUMBER_PASSWORD
}

func (n *NumberPassword) Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != n.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	userBasic, err := n.repo.GetUserBasicByNumber(ctx, req.Identifier)
	if err != nil {
		return nil, err
	}
	return checkPasswordAndGenToken(ctx, req, userBasic, n.jwtService, n.limiter, n.passHasher)
}
