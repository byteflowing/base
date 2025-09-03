package user

import (
	"context"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/base/ecode"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
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
	tx *query.Query,
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
				Data: &userv1.SignInResp_Data{
					Rule: rule,
				},
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
	accessToken, refreshToken, _, _, err := jwtService.GenerateToken(ctx, tx, &GenerateJwtReq{
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
			UserInfo:     userBasicToUserInfo(userBasic),
		},
	}
	return resp, nil
}

func signOutBySessionId(
	ctx context.Context,
	req *userv1.SignOutReq,
	repo Repo,
	tx *query.Query,
	jwtService *JwtService,
) (*userv1.SignOutResp, error) {
	log, err := repo.GetSignInLogByAccess(ctx, tx, req.SessionId)
	if err != nil {
		return nil, err
	}
	if err := jwtService.RevokeByLog(ctx, log); err != nil {
		return nil, err
	}
	log.Status = int16(req.Reason)
	err = repo.UpdateSignInLogByID(ctx, tx, log)
	return nil, err
}

func checkUserBasicUnique(
	ctx context.Context,
	tx *query.Query,
	repo Repo,
	biz string,
	phone *commonv1.PhoneNumber,
	number *string,
	email *string,
) (err error) {
	if phone != nil {
		exist, err := repo.CheckPhoneExists(ctx, tx, biz, phone)
		if err != nil {
			return err
		}
		if exist {
			return ecode.ErrPhoneExists
		}
	}
	if number != nil && *number != "" {
		exist, err := repo.CheckUserNumberExists(ctx, tx, *number)
		if err != nil {
			return err
		}
		if exist {
			return ecode.ErrUserNumberExists
		}
	}
	if email != nil && *email != "" {
		exist, err := repo.CheckEmailExists(ctx, tx, biz, *email)
		if err != nil {
			return err
		}
		if exist {
			return ecode.ErrEmailExists
		}
	}
	return nil
}
