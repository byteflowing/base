package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/go-common/crypto"
)

type NamePassword struct {
	passHasher crypto.PasswordHasher
	repo       Repo
	jwtService *JwtService
}

func (u *NamePassword) AuthType() AuthType {
	return AuthTypeNamePassword
}

func (u *NamePassword) Authenticate(ctx context.Context, req *SignInReq) (resp *SignInResp, err error) {
	if req.AuthType != AuthTypeNamePassword {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	userBasic, err := u.repo.GetUserBasicByName(ctx, req.Identifier)
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
	ok, err := u.passHasher.VerifyPassword(req.Credential, *userBasic.Password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ecode.ErrUserPasswordMisMatch
	}
	// 生成jwt token
	accessToken, refreshToken, err := u.jwtService.GenerateToken(ctx, userBasic, req)
	if err != nil {
		return nil, err
	}
	resp = &SignInResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
}
