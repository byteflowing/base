package user

import (
	"context"
	"time"

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
			UserInfo:     userBasicToUserInfo(userBasic),
		},
	}
	return resp, nil
}

func signOutBySessionId(
	ctx context.Context,
	req *userv1.SignOutReq,
	repo Repo,
	jwtService *JwtService,
) (*userv1.SignOutResp, error) {
	log, err := repo.GetSignInLogByAccess(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}
	if err := jwtService.RevokeByLog(ctx, log); err != nil {
		return nil, err
	}
	log.Status = int16(req.Reason)
	err = repo.UpdateSignInLogByID(ctx, log)
	return nil, err
}

func signOutByUid(
	ctx context.Context,
	uid int64,
	status enumsv1.SessionStatus,
	repo Repo,
	jwtService *JwtService,
) (err error) {
	logs, err := repo.GetActiveSignInLogs(ctx, uid)
	if err != nil {
		return err
	}
	if len(logs) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(logs))
	sessionItems := make([]*SessionItem, 0, len(logs)*2)
	for idx, log := range logs {
		ids[idx] = log.ID
		sessionItems[idx] = &SessionItem{
			SessionID: log.AccessSessionID,
			TTL:       time.Duration(log.AccessExpiredAt) * time.Second,
		}
		sessionItems[idx+1] = &SessionItem{
			SessionID: log.RefreshSessionID,
			TTL:       time.Duration(log.RefreshExpiredAt) * time.Second,
		}
	}
	if err := jwtService.RevokeTokens(ctx, sessionItems); err != nil {
		return err
	}
	return repo.UpdateSignInLogsStatus(ctx, ids, status)
}

// 唯一性校验，使用数据库的唯一索引兜底，这里就不使用分布式锁了
func checkUserBasicUnique(
	ctx context.Context,
	req *userv1.SignUpReq,
	repo Repo,
) (err error) {
	if req.Phone != nil {
		exist, err := repo.CheckPhoneExists(ctx, req.Phone)
		if err != nil {
			return err
		}
		if exist {
			return ecode.ErrPhoneExists
		}
	}
	if req.Number != nil && *req.Number != "" {
		exist, err := repo.CheckUserNumberExists(ctx, *req.Number)
		if err != nil {
			return err
		}
		if exist {
			return ecode.ErrUserNumberExists
		}
	}
	if req.Email != nil && *req.Email != "" {
		exist, err := repo.CheckEmailExists(ctx, *req.Email)
		if err != nil {
			return err
		}
		if exist {
			return ecode.ErrEmailExists
		}
	}
	return nil
}
