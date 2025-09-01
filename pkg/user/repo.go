package user

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/dal/query"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
)

type Repo interface {
	Transaction(fn func(tx *query.Query) error) error
	GetUserBasicByNumber(ctx context.Context, number string) (basic *model.UserBasic, err error)
	GetUserBasicByPhone(ctx context.Context, biz string, phone *commonv1.PhoneNumber) (basic *model.UserBasic, err error)
	GetUserBasicByEmail(ctx context.Context, biz string, email string) (basic *model.UserBasic, err error)
	GetUserBasicByUID(ctx context.Context, uid int64) (basic *model.UserBasic, err error)
	GetUserAuthByOpenID(ctx context.Context, openid string) (auth *model.UserAuth, err error)
	GetUserAuthByUnionID(ctx context.Context, biz, unionid string) (auth *model.UserAuth, err error)
	GetUidByPhone(ctx context.Context, biz string, phone *commonv1.PhoneNumber) (uid int64, err error)
	GetUidByEmail(ctx context.Context, biz string, email string) (uid int64, err error)
	CreateUserBasic(ctx context.Context, user *model.UserBasic) (err error)
	CreateUserAuth(ctx context.Context, user *model.UserAuth) (err error)
	UpdateUserAuth(ctx context.Context, auth *model.UserAuth) (err error)
	AddSignInLog(ctx context.Context, req *userv1.SignInReq, accessClaims, refreshClaims *JwtClaims) (err error)
	GetSignInLogByAccess(ctx context.Context, accessSessionID string) (log *model.UserSignLog, err error)
	GetSignInLogByRefresh(ctx context.Context, refreshSessionID string) (log *model.UserSignLog, err error)
	GetActiveSignInLogs(ctx context.Context, uid int64) (logs []*model.UserSignLog, err error)
	UpdateSignInLogByID(ctx context.Context, log *model.UserSignLog) (err error)
	UpdateSignInLogsStatus(ctx context.Context, ids []int64, status enumsv1.SessionStatus) (err error)
	CheckUserNumberExists(ctx context.Context, number string) (exist bool, err error)
	CheckEmailExists(ctx context.Context, biz string, email string) (exist bool, err error)
	CheckPhoneExists(ctx context.Context, biz string, number *commonv1.PhoneNumber) (exist bool, err error)
}

type GenRepo struct {
	db    *query.Query
	cache Cache
}

func NewStore(db *query.Query, cache Cache) *GenRepo {
	return &GenRepo{
		db:    db,
		cache: cache,
	}
}

func (repo *GenRepo) Transaction(fn func(tx *query.Query) error) error {
	return repo.db.Transaction(fn)
}

func (repo *GenRepo) GetUserBasicByNumber(ctx context.Context, number string) (basic *model.UserBasic, err error) {
	q := repo.db.UserBasic
	basic, err = q.WithContext(ctx).Where(q.Number.Eq(number)).Take()
	if err != nil {
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUserBasicByPhone(ctx context.Context, biz string, phone *commonv1.PhoneNumber) (basic *model.UserBasic, err error) {
	q := repo.db.UserBasic
	basic, err = q.WithContext(ctx).
		Where(
			q.Biz.Eq(biz),
			q.CountryCode.Eq(phone.CountryCode),
			q.Phone.Eq(phone.Number),
		).Take()
	if err != nil {
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUserBasicByEmail(ctx context.Context, biz string, email string) (basic *model.UserBasic, err error) {
	q := repo.db.UserBasic
	basic, err = q.WithContext(ctx).Where(
		q.Biz.Eq(biz),
		q.Email.Eq(email),
	).Take()
	if err != nil {
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUidByPhone(ctx context.Context, biz string, phone *commonv1.PhoneNumber) (uid int64, err error) {
	q := repo.db.UserBasic
	if err = q.WithContext(ctx).Select(q.ID).Where(
		q.Biz.Eq(biz),
		q.PhoneCountryCode.Eq(phone.CountryCode),
		q.Phone.Eq(phone.Number),
	).Pluck(q.ID, &uid); err != nil {
		return 0, err
	}
	return uid, nil
}

func (repo *GenRepo) GetUidByEmail(ctx context.Context, biz string, email string) (uid int64, err error) {
	q := repo.db.UserBasic
	if err = q.WithContext(ctx).Select(q.ID).Where(
		q.Biz.Eq(biz),
		q.Email.Eq(email),
	).Pluck(q.ID, &uid); err != nil {
		return 0, err
	}
	return uid, nil
}

func (repo *GenRepo) GetUserBasicByUID(ctx context.Context, uid int64) (basic *model.UserBasic, err error) {
	q := repo.db.UserBasic
	basic, err = q.WithContext(ctx).Where(q.ID.Eq(uid)).Take()
	if err != nil {
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUserAuthByOpenID(ctx context.Context, openid string) (auth *model.UserAuth, err error) {
	authQ := repo.db.UserAuth
	auth, err = authQ.WithContext(ctx).Where(authQ.Identifier.Eq(openid)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUserAuthByUnionID(ctx context.Context, biz, unionid string) (auth *model.UserAuth, err error) {
	authQ := repo.db.UserAuth
	auth, err = authQ.WithContext(ctx).Where(
		authQ.Biz.Eq(biz),
		authQ.UnionID.Eq(unionid),
	).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return
}

func (repo *GenRepo) CreateUserBasic(ctx context.Context, user *model.UserBasic) (err error) {
	q := repo.db.UserBasic
	return q.WithContext(ctx).Create(user)
}

func (repo *GenRepo) CreateUserAuth(ctx context.Context, user *model.UserAuth) (err error) {
	q := repo.db.UserAuth
	return q.WithContext(ctx).Create(user)
}

func (repo *GenRepo) UpdateUserAuth(ctx context.Context, auth *model.UserAuth) (err error) {
	q := repo.db.UserAuth
	_, err = q.WithContext(ctx).Where(q.ID.Eq(auth.ID)).Updates(auth)
	return
}

func (repo *GenRepo) CheckPhoneExists(ctx context.Context, biz string, number *commonv1.PhoneNumber) (exist bool, err error) {
	q := repo.db.UserBasic
	_, err = q.WithContext(ctx).Select(q.ID).Where(
		q.Biz.Eq(biz),
		q.PhoneCountryCode.Eq(number.CountryCode),
		q.Phone.Eq(number.Number)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo *GenRepo) CheckEmailExists(ctx context.Context, biz string, email string) (exist bool, err error) {
	q := repo.db.UserBasic
	_, err = q.WithContext(ctx).Select(q.ID).
		Where(
			q.Biz.Eq(biz),
			q.Email.Eq(email),
		).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo *GenRepo) CheckUserNumberExists(ctx context.Context, number string) (exist bool, err error) {
	q := repo.db.UserBasic
	_, err = q.WithContext(ctx).Select(q.ID).Where(q.Number.Eq(number)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo *GenRepo) AddSignInLog(ctx context.Context, req *userv1.SignInReq, accessClaims, refreshClaims *JwtClaims) (err error) {
	return repo.db.UserSignLog.WithContext(ctx).Create(&model.UserSignLog{
		UID:              int64(accessClaims.Uid),
		Type:             int16(req.AuthType),
		Status:           int16(enumsv1.SessionStatus_SESSION_STATUS_OK),
		Identifier:       repo.getIdentifier(req),
		Biz:              req.Biz,
		IP:               req.Ip,
		Location:         req.Location,
		Agent:            req.UserAgent,
		Device:           req.Device,
		AccessSessionID:  accessClaims.ID,
		RefreshSessionID: refreshClaims.ID,
		AccessExpiredAt:  accessClaims.ExpiresAt.Unix(),
		RefreshExpiredAt: refreshClaims.ExpiresAt.Unix(),
	})
}

// GetActiveSignInLogs 获取更新token没有过期的记录
// DESC排序
func (repo *GenRepo) GetActiveSignInLogs(ctx context.Context, uid int64) (logs []*model.UserSignLog, err error) {
	q := repo.db.UserSignLog
	now := time.Now().UnixMilli()
	logs, err = q.WithContext(ctx).Where(
		q.UID.Eq(int64(uid)),
		q.Status.Eq(int16(enumsv1.SessionStatus_SESSION_STATUS_OK)),
		q.RefreshExpiredAt.Gt(now),
	).Order(q.RefreshExpiredAt.Desc()).Find()
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

func (repo *GenRepo) UpdateSignInLogByID(ctx context.Context, log *model.UserSignLog) (err error) {
	q := repo.db.UserSignLog
	_, err = q.WithContext(ctx).Where(q.ID.Eq(log.ID)).Updates(log)
	return
}

func (repo *GenRepo) UpdateSignInLogsStatus(ctx context.Context, ids []int64, status enumsv1.SessionStatus) (err error) {
	q := repo.db.UserSignLog
	_, err = q.WithContext(ctx).Where(q.ID.In(ids...)).Update(q.Status, int16(status))
	return
}

func (repo *GenRepo) getIdentifier(req *userv1.SignInReq) (identifier string) {
	if req.AuthType == enumsv1.AuthType_AUTH_TYPE_PHONE_CAPTCHA || req.AuthType == enumsv1.AuthType_AUTH_TYPE_PHONE_PASSWORD {
		return req.PhoneNumber.CountryCode + req.PhoneNumber.Number
	}
	return req.Identifier
}
