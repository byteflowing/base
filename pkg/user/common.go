package user

import (
	"context"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/go-common/crypto"
)

func isDisabled(userBasic *model.UserBasic) bool {
	return userBasic.Status == int16(enumsv1.UserStatus_USER_STATUS_DISABLED)
}

func isAuthDisabled(userAuth *model.UserAuth) bool {
	return userAuth.Status == int16(enumsv1.AuthStatus_AUTH_STATUS_DISABLED)
}

func checkPasswordAndGenToken(
	ctx context.Context,
	req *userv1.SignInReq,
	userBasic *model.UserBasic,
	jwtService *JwtService,
	limiter Limiter,
	passHasher *crypto.PasswordHasher,
) (resp *userv1.SignInResp, err error) {
	if isDisabled(userBasic) {
		return nil, ecode.ErrUserDisabled
	}
	if req.AuthType == enumsv1.AuthType_AUTH_TYPE_NUMBER_PASSWORD ||
		req.AuthType == enumsv1.AuthType_AUTH_TYPE_PHONE_PASSWORD ||
		req.AuthType == enumsv1.AuthType_AUTH_TYPE_EMAIL_PASSWORD {
		// 验证密码是否正确
		if userBasic.Password == nil {
			return nil, ecode.ErrUserPasswordNotSet
		}
		// 检查密码错误次数
		rule, allow, err := limiter.AllowErr(ctx, userBasic.ID)
		if err != nil {
			return nil, err
		}
		if !allow {
			resp = &userv1.SignInResp{
				Rule: rule,
			}
			return resp, nil
		}
		ok, err := passHasher.VerifyPassword(req.Credential, *userBasic.Password)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ecode.ErrUserPasswordMisMatch
		}
		if err = limiter.ResetErr(ctx, userBasic.ID); err != nil {
			return nil, err
		}
	}
	// 生成jwt token
	accessToken, refreshToken, _, _, err := jwtService.GenerateToken(ctx, &GenerateJwtReq{
		UserBasic:      userBasic,
		SignInReq:      req,
		ExtraJwtClaims: req.ExtraJwtClaims,
		AuthType:       req.AuthType,
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
