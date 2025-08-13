package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/go-common/crypto"
)

type EmailPassword struct {
	passHasher crypto.PasswordHasher
	repo       Repo
	jwtService *JwtService
}

func NewEmailPassword(passHasher crypto.PasswordHasher, repo Repo, jwtService *JwtService) Authenticator {
	return &EmailPassword{
		passHasher: passHasher,
		repo:       repo,
		jwtService: jwtService,
	}
}

func (e *EmailPassword) AuthType() AuthType {
	return AuthTypeEmailPassword
}

func (e *EmailPassword) Authenticate(ctx context.Context, req *SignInReq) (resp *SignInResp, err error) {
	if req.AuthType != AuthTypeEmailPassword {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	userBasic, err := e.repo.GetUserBasicByEmail(ctx, req.Identifier)
	if err != nil {
		return nil, err
	}
	// 检查用户是否被禁用
	if Status(userBasic.Status) == StatusDisabled {
		return nil, ecode.ErrUserDisabled
	}
	// 验证密码是否正确
	if userBasic.Password == nil {
		return nil, ecode.ErrUserPasswordNotSet
	}
	ok, err := e.passHasher.VerifyPassword(req.Credential, *userBasic.Password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ecode.ErrUserPasswordMisMatch
	}
	// 生成jwt token
	accessToken, refreshToken, err := e.jwtService.GenerateToken(ctx, userBasic, req)
	if err != nil {
		return nil, err
	}
	resp = &SignInResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
}
