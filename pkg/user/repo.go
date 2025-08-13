package user

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/go-common/trans"
)

type Repo interface {
	GetUserBasicByName(ctx context.Context, name string) (basic *model.UserBasic, err error)
	GetUserBasicByPhone(ctx context.Context, phone string) (basic *model.UserBasic, err error)
	GetUserBasicByEmail(ctx context.Context, email string) (basic *model.UserBasic, err error)
	GetOrCreateUserAuthByWechat(ctx context.Context, openid, unionid, sessionKey string) (userAuth *model.UserBasic, err error)
	AddSignInLog(ctx context.Context, req *SignInReq, accessClaims, refreshClaims *JwtClaims) (err error)
	RefreshSignInLog(ctx context.Context, oldRefreshSessionID string, accessClaims, refreshClaims *JwtClaims) (err error)
	DisActiveSignInLogByAccess(ctx context.Context, AccessSessionID string, status SessionStatus) (err error)
	DisActiveSignInLogByRefresh(ctx context.Context, RefreshSessionID string, status SessionStatus) (err error)
	GetSignInLogByAccess(ctx context.Context, accessSessionID string) (log *model.UserSignLog, err error)
	GetSignInLogByRefresh(ctx context.Context, refreshSessionID string) (log *model.UserSignLog, err error)
	ExpireAccessSessionID(ctx context.Context, accessSessionID string) (err error)
}

type GenRepo struct {
	db      *query.Query
	cache   Cache
	shortID *common.ShortIDGenerator
}

func NewStore(db *query.Query, cache Cache) *GenRepo {
	return &GenRepo{
		db:    db,
		cache: cache,
	}
}

func (repo *GenRepo) GetUserBasicByName(ctx context.Context, name string) (basic *model.UserBasic, err error) {
	q := repo.db.UserBasic
	basic, err = q.WithContext(ctx).Where(q.Name.Eq(name)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrUserNotExist
		}
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUserBasicByPhone(ctx context.Context, phone string) (basic *model.UserBasic, err error) {
	q := repo.db.UserBasic
	basic, err = q.WithContext(ctx).Where(q.Phone.Eq(phone)).Take()
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

func (repo *GenRepo) GetOrCreateUserAuthByWechat(ctx context.Context, openid, unionid, sessionKey string) (userBasic *model.UserBasic, err error) {
	err = repo.db.Transaction(func(tx *query.Query) error {
		auth, err := tx.UserAuth.WithContext(ctx).Where(tx.UserAuth.Identifier.Eq(openid)).Take()
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		// 如果没有auth就新建一个用户
		if err != nil {
			userBasic = &model.UserBasic{
				Number: trans.Ref(repo.shortID.GetID()),
				Name:   repo.shortID.GetID(),
				Alias_: repo.shortID.GetID(),
				Status: StatusOk,
				Source: int16(SourceWechat),
			}
			if err := tx.UserBasic.WithContext(ctx).Create(userBasic); err != nil {
				return err
			}
			userAuth := &model.UserAuth{
				UID:        userBasic.ID,
				Type:       int16(AuthTypeWechat),
				Status:     userBasic.Status,
				Identifier: openid,
				Credential: sessionKey,
				UnionID:    trans.Ref(unionid),
			}
			if err := tx.UserAuth.WithContext(ctx).Create(userAuth); err != nil {
				return err
			}
			return nil
		} else {
			// 如果openid匹配就更新
			// 更新unionid，如果账号绑定了微信开放平台就可以关联上
			if auth.Status == int16(AuthStatusDisabled) {
				return ecode.ErrUserAuthDisabled
			}
			m := &model.UserAuth{Credential: sessionKey}
			if unionid != "" && auth.UnionID == nil {
				m.UnionID = trans.Ref(unionid)
			}
			if _, err = tx.UserAuth.WithContext(ctx).Where(tx.UserAuth.ID.Eq(auth.ID)).Updates(m); err != nil {
				return err
			}
			if userBasic, err = tx.UserBasic.WithContext(ctx).Where(tx.UserBasic.ID.Eq(auth.UID)).Take(); err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func (repo *GenRepo) AddSignInLog(ctx context.Context, req *SignInReq, accessClaims, refreshClaims *JwtClaims) (err error) {
	return repo.db.UserSignLog.WithContext(ctx).Create(&model.UserSignLog{
		UID:              int64(accessClaims.Uid),
		AccessSessionID:  accessClaims.ID,
		RefreshSessionID: refreshClaims.ID,
		Type:             int16(req.AuthType),
		Status:           int16(SessionStatusSignIn),
		IP:               req.IP,
		Location:         req.Location,
		Agent:            req.UserAgent,
		Device:           req.Device,
		AccessExpiredAt:  accessClaims.ExpiresAt.UnixMilli(),
		RefreshExpiredAt: refreshClaims.ExpiresAt.UnixMilli(),
	})
}

func (repo *GenRepo) RefreshSignInLog(ctx context.Context, oldRefreshSessionID string, accessClaims, refreshClaims *JwtClaims) (err error) {
	q := repo.db.UserSignLog
	_, err = q.WithContext(ctx).Where(q.RefreshSessionID.Eq(oldRefreshSessionID)).Updates(model.UserSignLog{
		AccessSessionID:  accessClaims.ID,
		RefreshSessionID: refreshClaims.ID,
		Status:           int16(SessionStatusSignIn),
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

func (repo *GenRepo) DisActiveSignInLogByAccess(ctx context.Context, AccessSessionID string, status SessionStatus) (err error) {
	q := repo.db.UserSignLog
	now := time.Now().UnixMilli()
	_, err = q.WithContext(ctx).Where(q.AccessSessionID.Eq(AccessSessionID)).Updates(&model.UserSignLog{
		Status:           int16(status),
		AccessExpiredAt:  now,
		RefreshExpiredAt: now,
	})
	return
}

func (repo *GenRepo) DisActiveSignInLogByRefresh(ctx context.Context, RefreshSessionID string, status SessionStatus) (err error) {
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
		q.Status.Eq(int16(SessionStatusSignIn)),
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
