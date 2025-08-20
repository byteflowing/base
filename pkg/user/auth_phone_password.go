package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/go-common/crypto"
)

type PhonePassword struct {
	passHasher crypto.PasswordHasher
	repo       Repo
	jwtService *JwtService
}

func NewPhonePassword(passHasher crypto.PasswordHasher, repo Repo, jwtService *JwtService) Authenticator {
	return &EmailPassword{
		passHasher: passHasher,
		repo:       repo,
		jwtService: jwtService,
	}
}

func (p *PhonePassword) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_PHONE_PASSWORD
}

func (p *PhonePassword) Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != enumsv1.AuthType_AUTH_TYPE_PHONE_PASSWORD {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.PhoneNumber == nil || req.PhoneNumber.Number == "" || req.PhoneNumber.CountryCode == "" {
		return nil, ecode.ErrPhoneIsEmpty
	}
	userBasic, err := p.repo.GetUserBasicByPhone(ctx, req.GetPhoneNumber())
	if err != nil {
		return nil, err
	}
	// 检查用户是否被禁用
	if isDisabled(userBasic) {
		return nil, ecode.ErrUserDisabled
	}
	// 验证密码是否正确
	if userBasic.Password == nil {
		return nil, ecode.ErrUserPasswordNotSet
	}
	ok, err := p.passHasher.VerifyPassword(req.Credential, *userBasic.Password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ecode.ErrUserPasswordMisMatch
	}
	// 生成jwt token
	accessToken, refreshToken, err := p.jwtService.GenerateToken(ctx, &GenerateJwtReq{
		UserBasic:      userBasic,
		SignInReq:      req,
		ExtraJwtClaims: req.ExtraJwtClaims,
		AuthType:       enumsv1.AuthType_AUTH_TYPE_PHONE_PASSWORD,
	})
	if err != nil {
		return nil, err
	}
	resp = &userv1.SignInResp{
		Data: &userv1.SignInResp_Data{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
	return resp, nil
}
