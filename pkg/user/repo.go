package user

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/base/ecode"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/go-common/trans"
)

type Repo interface {
	GetUserBasicByNumber(ctx context.Context, number string) (basic *model.UserBasic, err error)
	GetUserBasicByPhone(ctx context.Context, phone *commonv1.PhoneNumber) (basic *model.UserBasic, err error)
	GetUserBasicByEmail(ctx context.Context, email string) (basic *model.UserBasic, err error)
	GetOrCreateUserAuthByWechat(ctx context.Context, appid, openid, unionid, sessionKey string) (userAuth *model.UserBasic, err error)
	AddSignInLog(ctx context.Context, req *userv1.SignInReq, accessClaims, refreshClaims *JwtClaims) (err error)
	RefreshSignInLog(ctx context.Context, oldRefreshSessionID string, accessClaims, refreshClaims *JwtClaims) (err error)
	DisActiveSignInLogByAccess(ctx context.Context, AccessSessionID string, status enumsv1.SessionStatus) (err error)
	DisActiveSignInLogByRefresh(ctx context.Context, RefreshSessionID string, status enumsv1.SessionStatus) (err error)
	GetSignInLogByAccess(ctx context.Context, accessSessionID string) (log *model.UserSignLog, err error)
	GetSignInLogByRefresh(ctx context.Context, refreshSessionID string) (log *model.UserSignLog, err error)
	ExpireAccessSessionID(ctx context.Context, accessSessionID string) (err error)
}

type GenRepo struct {
	db          *query.Query
	cache       Cache
	shortIDGen  *common.ShortIDGenerator
	globalIDGen common.GlobalIdGenerator
}

func NewStore(db *query.Query, cache Cache, globalIDGen common.GlobalIdGenerator, shortIDGen *common.ShortIDGenerator) *GenRepo {
	return &GenRepo{
		db:          db,
		cache:       cache,
		shortIDGen:  shortIDGen,
		globalIDGen: globalIDGen,
	}
}

func (repo *GenRepo) GetUserBasicByNumber(ctx context.Context, number string) (basic *model.UserBasic, err error) {
	q := repo.db.UserBasic
	basic, err = q.WithContext(ctx).Where(q.Number.Eq(number)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrUserNotExist
		}
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUserBasicByPhone(ctx context.Context, phone *commonv1.PhoneNumber) (basic *model.UserBasic, err error) {
	q := repo.db.UserBasic
	basic, err = q.WithContext(ctx).
		Where(q.CountryCode.Eq(phone.CountryCode), q.Phone.Eq(phone.Number)).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrUserNotExist
		}
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUserBasicByEmail(ctx context.Context, email string) (basic *model.UserBasic, err error) {
	q := repo.db.UserBasic
	basic, err = q.WithContext(ctx).Where(q.Email.Eq(email)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrUserNotExist
		}
		return nil, err
	}
	return
}

func (repo *GenRepo) GetOrCreateUserAuthByWechat(ctx context.Context, appid, openid, unionid, sessionKey string) (userBasic *model.UserBasic, err error) {
	err = repo.db.Transaction(func(tx *query.Query) error {
		auth, err := tx.UserAuth.WithContext(ctx).Where(tx.UserAuth.Identifier.Eq(openid)).Take()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// 如果没有auth就新建一个用户
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 先尝试通过unionid查找用户
			if unionid != "" {
				if ua, err := tx.UserAuth.WithContext(ctx).Select(tx.UserAuth.UID).Where(tx.UserAuth.UnionID.Eq(unionid)).Take(); err == nil {
					// err == nil 找到用户
					userBasic, err = tx.UserBasic.WithContext(ctx).Where(tx.UserBasic.ID.Eq(ua.UID)).Take()
					if err != nil {
						return err
					}
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
			}
			// 如果没有找到用户就新建
			uid, err := repo.globalIDGen.GetID()
			if err != nil {
				return err
			}
			number, err := repo.shortIDGen.GetID()
			if err != nil {
				return err
			}
			if userBasic == nil {
				userBasic = &model.UserBasic{
					ID:     uid,
					Number: number,
					Status: int16(enumsv1.UserStatus_USER_STATUS_OK),
					Source: int16(enumsv1.UserSource_USER_SOURCE_WECHAT),
				}
				if err := tx.UserBasic.WithContext(ctx).Create(userBasic); err != nil {
					return err
				}
			}
			newAuth := &model.UserAuth{
				UID:        userBasic.ID,
				Type:       int16(enumsv1.AuthType_AUTH_TYPE_WECHAT),
				Status:     int16(enumsv1.AuthStatus_AUTH_STATUS_OK),
				Appid:      appid,
				Identifier: openid,
				Credential: sessionKey,
				UnionID:    trans.Ref(unionid),
			}
			return tx.UserAuth.WithContext(ctx).Create(newAuth)
		}

		// 如果openid匹配就更新
		// 更新unionid，如果账号绑定了微信开放平台就可以关联上
		if auth.Status == int16(enumsv1.AuthStatus_AUTH_STATUS_DISABLED) {
			return ecode.ErrUserAuthDisabled
		}
		m := &model.UserAuth{Credential: sessionKey}
		if unionid != "" && auth.UnionID == nil {
			m.UnionID = trans.Ref(unionid)
		}
		if _, err = tx.UserAuth.WithContext(ctx).Where(tx.UserAuth.ID.Eq(auth.ID)).Updates(m); err != nil {
			return err
		}
		userBasic, err = tx.UserBasic.WithContext(ctx).Where(tx.UserBasic.ID.Eq(auth.UID)).Take()
		return err
	})
	return
}

func (repo *GenRepo) AddSignInLog(ctx context.Context, req *userv1.SignInReq, accessClaims, refreshClaims *JwtClaims) (err error) {
	return repo.db.UserSignLog.WithContext(ctx).Create(&model.UserSignLog{
		UID:              int64(accessClaims.Uid),
		Type:             int16(req.AuthType),
		Status:           int16(enumsv1.SessionStatus_SESSION_STATUS_OK),
		Identifier:       repo.getIdentifier(req),
		IP:               req.Ip,
		Location:         req.Location,
		Agent:            req.UserAgent,
		Device:           req.Device,
		AccessSessionID:  accessClaims.ID,
		RefreshSessionID: refreshClaims.ID,
		AccessExpiredAt:  accessClaims.ExpiresAt.UnixMilli(),
		RefreshExpiredAt: refreshClaims.ExpiresAt.UnixMilli(),
	})
}

func (repo *GenRepo) RefreshSignInLog(ctx context.Context, oldRefreshSessionID string, accessClaims, refreshClaims *JwtClaims) (err error) {
	q := repo.db.UserSignLog
	_, err = q.WithContext(ctx).Where(q.RefreshSessionID.Eq(oldRefreshSessionID)).Updates(model.UserSignLog{
		AccessSessionID:  accessClaims.ID,
		RefreshSessionID: refreshClaims.ID,
		Status:           int16(enumsv1.SessionStatus_SESSION_STATUS_OK),
		AccessExpiredAt:  accessClaims.ExpiresAt.UnixMilli(),
		RefreshExpiredAt: refreshClaims.ExpiresAt.UnixMilli(),
	})
	return
}

func (repo *GenRepo) ExpireAccessSessionID(ctx context.Context, accessSessionID string) (err error) {
	q := repo.db.UserSignLog
	now := time.Now().UnixMilli()
	_, err = q.WithContext(ctx).Where(q.AccessSessionID.Eq(accessSessionID)).Updates(model.UserSignLog{
		AccessExpiredAt: now,
	})
	return
}

func (repo *GenRepo) DisActiveSignInLogByAccess(ctx context.Context, AccessSessionID string, status enumsv1.SessionStatus) (err error) {
	q := repo.db.UserSignLog
	now := time.Now().UnixMilli()
	_, err = q.WithContext(ctx).Where(q.AccessSessionID.Eq(AccessSessionID)).Updates(&model.UserSignLog{
		Status:           int16(status),
		AccessExpiredAt:  now,
		RefreshExpiredAt: now,
	})
	return
}

func (repo *GenRepo) DisActiveSignInLogByRefresh(ctx context.Context, RefreshSessionID string, status enumsv1.SessionStatus) (err error) {
	q := repo.db.UserSignLog
	now := time.Now().UnixMilli()
	_, err = q.WithContext(ctx).Where(q.RefreshSessionID.Eq(RefreshSessionID)).Updates(&model.UserSignLog{
		Status:           int16(status),
		AccessExpiredAt:  now,
		RefreshExpiredAt: now,
	})
	return
}

func (repo *GenRepo) GetActiveSignInLog(ctx context.Context, uid uint64) (logs []*model.UserSignLog, err error) {
	q := repo.db.UserSignLog
	now := time.Now().UnixMilli()
	logs, err = q.WithContext(ctx).Where(
		q.UID.Eq(int64(uid)),
		q.Status.Eq(int16(enumsv1.SessionStatus_SESSION_STATUS_OK)),
		q.RefreshExpiredAt.Gt(now),
	).Find()
	return
}

func (repo *GenRepo) GetSignInLogByAccess(ctx context.Context, accessSessionID string) (log *model.UserSignLog, err error) {
	q := repo.db.UserSignLog
	log, err = q.WithContext(ctx).Where(q.AccessSessionID.Eq(accessSessionID)).Take()
	return
}

func (repo *GenRepo) GetSignInLogByRefresh(ctx context.Context, refreshSessionID string) (log *model.UserSignLog, err error) {
	q := repo.db.UserSignLog
	log, err = q.WithContext(ctx).Where(q.RefreshSessionID.Eq(refreshSessionID)).Take()
	return
}

func (repo *GenRepo) getIdentifier(req *userv1.SignInReq) (identifier string) {
	if req.AuthType == enumsv1.AuthType_AUTH_TYPE_PHONE_CAPTCHA || req.AuthType == enumsv1.AuthType_AUTH_TYPE_PHONE_PASSWORD {
		return req.PhoneNumber.CountryCode + req.PhoneNumber.Number
	}
	return req.Identifier
}
