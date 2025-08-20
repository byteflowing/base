package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
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

func (e *EmailPassword) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_EMAIL_PASSWORD
}

func (e *EmailPassword) Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != enumsv1.AuthType_AUTH_TYPE_EMAIL_PASSWORD {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	userBasic, err := e.repo.GetUserBasicByEmail(ctx, req.Identifier)
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
	ok, err := e.passHasher.VerifyPassword(req.Credential, *userBasic.Password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ecode.ErrUserPasswordMisMatch
	}
	// 生成jwt token
	accessToken, refreshToken, err := e.jwtService.GenerateToken(ctx, &GenerateJwtReq{
		UserBasic:      userBasic,
		SignInReq:      req,
		ExtraJwtClaims: req.ExtraJwtClaims,
		AuthType:       e.AuthType(),
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
