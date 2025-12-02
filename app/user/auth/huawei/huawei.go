package huawei

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/byteflowing/base/app/user/common"
	"github.com/byteflowing/base/app/user/dal/model"
	"github.com/byteflowing/base/app/user/dal/query"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/redis"
	"github.com/byteflowing/base/pkg/sdk/huawei/account"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	huaweiv1 "github.com/byteflowing/proto/gen/go/huawei/v1"
	userv1 "github.com/byteflowing/proto/gen/go/user/v1"
)

type AccountManager struct {
	appTokenKey string
	clientID    string
	idService   *common.IDService
	cli         *account.Account
	rdb         *redis.Redis
	lock        *redis.Lock
}

func NewAccountManager(
	keyPrefix string,
	rdb *redis.Redis,
	idService *common.IDService,
	config *huaweiv1.AccountConfig,
) *AccountManager {
	cli := account.New(config.ClientId, config.ClientSecret)
	return &AccountManager{
		rdb: rdb,
		cli: cli,
		lock: redis.NewLock(rdb, &redis.LockOption{
			Prefix: keyPrefix + ":hw_app_token:lock",
			Tries:  3,
			TTL:    30 * time.Second,
			Wait:   10 * time.Millisecond,
		}),
		idService:   idService,
		appTokenKey: keyPrefix + ":hw_app_token:" + config.ClientId,
		clientID:    config.ClientId,
	}
}

func (am *AccountManager) Authenticate(ctx context.Context, req *userv1.SignInReq, tx *query.Query) (*userv1.SignInResult, error) {
	if req == nil || req.SignInType != enumsv1.SignInType_SIGN_IN_TYPE_HUAWEI {
		return nil, errors.New("invalid params")
	}
	param := req.GetHuawei()
	result, err := am.cli.SignIn(ctx, param)
	if err != nil {
		return nil, err
	}
	user, needAuth, err := am.getUser(ctx, result, tx)
	if err != nil {
		return nil, err
	}
	if user == nil {
		number, err := am.idService.GetShortID(ctx)
		if err != nil {
			return nil, err
		}
		id, err := am.idService.GetGlobalID(ctx)
		if err != nil {
			return nil, err
		}
		agent := req.GetAgent()
		if agent == nil {
			agent = &userv1.Agent{}
		}
		user = &model.UserAccount{
			ID:               id,
			TenantID:         req.GetTenantId(),
			Number:           number,
			PhoneCountryCode: am.parseCountryCode(result.PhoneCountryCode),
			Phone:            result.PurePhoneNumber,
			Status:           int16(enumsv1.UserStatus_USER_STATUS_OK),
			Source:           int16(enumsv1.UserSource_USER_SOURCE_APP_HARMONY),
			SignupType:       int16(enumsv1.SignUpType_SIGN_UP_TYPE_HUAWEI_ACCOUNT),
			PhoneVerified:    true,
			RegisterIP:       agent.Ip,
			RegisterDevice:   agent.Device,
			RegisterAgent:    agent.Agent,
			RegisterLocation: common.LocationToString(agent.Location),
		}
		if err := tx.UserAccount.WithContext(ctx).Create(user); err != nil {
			return nil, err
		}
	} else {
		if !common.IsUserValid(user.Status) {
			return nil, ecode.ErrUserDisabled
		}
	}
	if needAuth {
		userAuth := &model.UserAuth{
			TenantID: req.GetTenantId(),
			UID:      user.ID,
			Type:     int16(enumsv1.SignInType_SIGN_IN_TYPE_HUAWEI),
			Status:   int16(enumsv1.UserStatus_USER_STATUS_OK),
			Appid:    am.clientID,
			OpenID:   result.OpenId,
			UnionID:  result.UnionId,
		}
		if err := tx.UserAuth.WithContext(ctx).Create(userAuth); err != nil {
			return nil, err
		}
	}
	return &userv1.SignInResult{
		User:       common.UserModelToUser(user),
		Identifier: result.OpenId,
	}, nil
}

func (am *AccountManager) getUser(ctx context.Context, result *huaweiv1.HuaweiSignInResp, tx *query.Query) (m *model.UserAccount, needCreateAuth bool, err error) {
	if err := am.parseError(result.ResultCode, 0, result.ResultDesc); err != nil {
		return nil, false, err
	}
	q := tx.UserAuth
	accountQ := tx.UserAccount
	userAuth, err := q.WithContext(ctx).Where(q.OpenID.Eq(result.OpenId), q.UnionID.Eq(result.UnionId)).Take()
	if err == nil {
		if userAuth.Status != int16(enumsv1.AuthStatus_AUTH_STATUS_OK) {
			return nil, false, ecode.ErrUserAuthInvalid
		}
		m, err = accountQ.WithContext(ctx).Where(accountQ.TenantID.Eq(userAuth.TenantID), accountQ.ID.Eq(userAuth.UID)).Take()
		return m, false, err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	if result.PhoneNumber == "" {
		return nil, true, nil
	}
	countryCode := am.parseCountryCode(result.PhoneCountryCode)
	userAccount, err := accountQ.WithContext(ctx).Where(accountQ.PhoneCountryCode.Eq(countryCode), accountQ.Phone.Eq(result.PurePhoneNumber)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, true, nil
		}
		return nil, false, err
	}
	return userAccount, true, nil
}

func (am *AccountManager) parseCountryCode(countryCode string) string {
	return strings.TrimLeft(countryCode, "00")
}

func (am *AccountManager) getAppToken(ctx context.Context) (token string, err error) {
	token, err = am.rdb.Get(ctx, am.appTokenKey).Result()
	if err == nil {
		return token, nil
	}
	if !errors.Is(err, redis.Nil) {
		return "", err
	}
	identifier, err := am.lock.Acquire(ctx, am.clientID)
	if err != nil {
		return "", err
	}
	defer am.lock.Release(ctx, am.clientID, identifier)
	if token, err := am.rdb.Get(ctx, am.appTokenKey).Result(); err == nil {
		return token, nil
	}
	resp, err := am.cli.GetAppToken(ctx, &huaweiv1.HuaweiGetAppTokenReq{})
	if err != nil {
		return "", err
	}
	if err := am.parseError(resp.Error, resp.SubError, resp.ErrorDescription); err != nil {
		return "", err
	}
	ttl := time.Duration(resp.ExpiresIn-10) * time.Second
	if err := am.rdb.Set(ctx, am.appTokenKey, resp.AccessToken, ttl).Err(); err != nil {
		return "", err
	}
	return resp.AccessToken, nil
}

func (am *AccountManager) parseError(err, subErr int32, desc string) error {
	if err != 0 {
		return fmt.Errorf("errCode: %d, subErrCode: %d, errMsg: %s", err, subErr, desc)
	}
	return nil
}
