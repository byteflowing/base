package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/go-common/crypto"
)

type PhonePassword struct {
	passHasher *crypto.PasswordHasher
	repo       Repo
	jwtService *JwtService
	limiter    Limiter
}

func NewPhonePassword(passHasher *crypto.PasswordHasher, repo Repo, jwtService *JwtService, limiter Limiter) Authenticator {
	return &PhonePassword{
		passHasher: passHasher,
		repo:       repo,
		jwtService: jwtService,
		limiter:    limiter,
	}
}

func (p *PhonePassword) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_PHONE_PASSWORD
}

func (p *PhonePassword) Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != p.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.PhoneNumber == nil || req.PhoneNumber.Number == "" || req.PhoneNumber.CountryCode == "" {
		return nil, ecode.ErrPhoneIsEmpty
	}
	userBasic, err := p.repo.GetUserBasicByPhone(ctx, req.GetPhoneNumber())
	if err != nil {
		return nil, err
	}
	return checkPasswordAndGenToken(ctx, req, userBasic, p.jwtService, p.limiter, p.passHasher)
}
