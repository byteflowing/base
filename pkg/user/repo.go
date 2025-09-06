package user

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/go-common/orm"
	"github.com/byteflowing/go-common/slicex"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
	typesv1 "github.com/byteflowing/proto/gen/go/types/v1"
)

type Repo interface {
	GetUserBasicByNumber(ctx context.Context, tx *query.Query, number string) (basic *model.UserBasic, err error)
	GetUserBasicByPhone(ctx context.Context, tx *query.Query, biz string, phone *typesv1.PhoneNumber) (basic *model.UserBasic, err error)
	GetUserBasicByEmail(ctx context.Context, tx *query.Query, biz string, email string) (basic *model.UserBasic, err error)
	GetUserBasicByUID(ctx context.Context, tx *query.Query, uid int64) (basic *model.UserBasic, err error)
	GetUserAuthByOpenID(ctx context.Context, tx *query.Query, openid string) (auth *model.UserAuth, err error)
	GetOneUserAuthByUnionID(ctx context.Context, tx *query.Query, biz, unionid string) (auth *model.UserAuth, err error)
	GetUidByPhone(ctx context.Context, tx *query.Query, biz string, phone *typesv1.PhoneNumber) (uid int64, err error)
	GetUidByEmail(ctx context.Context, tx *query.Query, biz string, email string) (uid int64, err error)
	CreateUserBasic(ctx context.Context, tx *query.Query, user *model.UserBasic) (err error)
	CreateUserAuth(ctx context.Context, tx *query.Query, user *model.UserAuth) (err error)
	UpdateUserAuth(ctx context.Context, tx *query.Query, auth *model.UserAuth) (err error)
	UpdateUserBasicByUid(ctx context.Context, tx *query.Query, basic *model.UserBasic) (err error)
	AddSignInLog(ctx context.Context, tx *query.Query, req *userv1.SignInReq, accessClaims, refreshClaims *JwtClaims) (err error)
	GetSignInLogByAccess(ctx context.Context, tx *query.Query, accessSessionID string) (log *model.UserSignLog, err error)
	GetSignInLogByRefresh(ctx context.Context, tx *query.Query, refreshSessionID string) (log *model.UserSignLog, err error)
	GetActiveSignInLogs(ctx context.Context, tx *query.Query, uid int64) (logs []*model.UserSignLog, err error)
	UpdateSignInLogByID(ctx context.Context, tx *query.Query, log *model.UserSignLog) (err error)
	UpdateSignInLogsStatus(ctx context.Context, tx *query.Query, ids []int64, status enumsv1.SessionStatus) (err error)
	CheckUserNumberExists(ctx context.Context, tx *query.Query, number string) (exist bool, err error)
	CheckEmailExists(ctx context.Context, tx *query.Query, biz string, email string) (exist bool, err error)
	CheckPhoneExists(ctx context.Context, tx *query.Query, biz string, number *typesv1.PhoneNumber) (exist bool, err error)
	PagingGetSignInLogs(ctx context.Context, tx *query.Query, req *userv1.PagingGetSignInLogsReq) (resp *orm.PageResult[model.UserSignLog], err error)
	PagingGetUsers(ctx context.Context, tx *query.Query, req *userv1.PagingGetUsersReq) (resp *orm.PageResult[model.UserBasic], err error)
	DeleteUserBasic(ctx context.Context, tx *query.Query, uid int64) error
	DeleteUsersBasic(ctx context.Context, tx *query.Query, uids []int64) error
	DeleteUserAuth(ctx context.Context, tx *query.Query, uid int64) error
	DeleteUsersAuth(ctx context.Context, tx *query.Query, uids []int64) error
	DeleteUserSignLogs(ctx context.Context, tx *query.Query, uid int64) error
	DeleteUsersSignLogs(ctx context.Context, tx *query.Query, uids []int64) error
	GetUserAuthByUid(ctx context.Context, tx *query.Query, uid int64) (auth []*model.UserAuth, err error)
	UnbindUserAuth(ctx context.Context, tx *query.Query, uid int64, id string) error
}

type GenRepo struct {
	cache Cache
}

func NewRepo(cache Cache) Repo {
	return &GenRepo{
		cache: cache,
	}
}

func (repo *GenRepo) GetUserBasicByNumber(ctx context.Context, tx *query.Query, number string) (basic *model.UserBasic, err error) {
	q := tx.UserBasic
	basic, err = q.WithContext(ctx).Where(q.Number.Eq(number)).Take()
	if err != nil {
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUserBasicByPhone(ctx context.Context, tx *query.Query, biz string, phone *typesv1.PhoneNumber) (basic *model.UserBasic, err error) {
	q := tx.UserBasic
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

func (repo *GenRepo) GetUserBasicByEmail(ctx context.Context, tx *query.Query, biz string, email string) (basic *model.UserBasic, err error) {
	q := tx.UserBasic
	basic, err = q.WithContext(ctx).Where(
		q.Biz.Eq(biz),
		q.Email.Eq(email),
	).Take()
	if err != nil {
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUidByPhone(ctx context.Context, tx *query.Query, biz string, phone *typesv1.PhoneNumber) (uid int64, err error) {
	q := tx.UserBasic
	if err = q.WithContext(ctx).Select(q.ID).Where(
		q.Biz.Eq(biz),
		q.PhoneCountryCode.Eq(phone.CountryCode),
		q.Phone.Eq(phone.Number),
	).Pluck(q.ID, &uid); err != nil {
		return 0, err
	}
	return uid, nil
}

func (repo *GenRepo) GetUidByEmail(ctx context.Context, tx *query.Query, biz string, email string) (uid int64, err error) {
	q := tx.UserBasic
	if err = q.WithContext(ctx).Select(q.ID).Where(
		q.Biz.Eq(biz),
		q.Email.Eq(email),
	).Pluck(q.ID, &uid); err != nil {
		return 0, err
	}
	return uid, nil
}

func (repo *GenRepo) GetUserBasicByUID(ctx context.Context, tx *query.Query, uid int64) (basic *model.UserBasic, err error) {
	q := tx.UserBasic
	basic, err = q.WithContext(ctx).Where(q.ID.Eq(uid)).Take()
	if err != nil {
		return nil, err
	}
	return
}

func (repo *GenRepo) GetUserAuthByOpenID(ctx context.Context, tx *query.Query, openid string) (auth *model.UserAuth, err error) {
	authQ := tx.UserAuth
	auth, err = authQ.WithContext(ctx).Where(authQ.Identifier.Eq(openid)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return
}

func (repo *GenRepo) GetOneUserAuthByUnionID(ctx context.Context, tx *query.Query, biz, unionid string) (auth *model.UserAuth, err error) {
	authQ := tx.UserAuth
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

func (repo *GenRepo) CreateUserBasic(ctx context.Context, tx *query.Query, user *model.UserBasic) (err error) {
	q := tx.UserBasic
	return q.WithContext(ctx).Create(user)
}

func (repo *GenRepo) CreateUserAuth(ctx context.Context, tx *query.Query, user *model.UserAuth) (err error) {
	q := tx.UserAuth
	return q.WithContext(ctx).Create(user)
}

func (repo *GenRepo) UpdateUserAuth(ctx context.Context, tx *query.Query, auth *model.UserAuth) (err error) {
	q := tx.UserAuth
	_, err = q.WithContext(ctx).Where(q.ID.Eq(auth.ID)).Updates(auth)
	return
}

func (repo *GenRepo) CheckPhoneExists(ctx context.Context, tx *query.Query, biz string, number *typesv1.PhoneNumber) (exist bool, err error) {
	q := tx.UserBasic
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

func (repo *GenRepo) CheckEmailExists(ctx context.Context, tx *query.Query, biz string, email string) (exist bool, err error) {
	q := tx.UserBasic
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

func (repo *GenRepo) CheckUserNumberExists(ctx context.Context, tx *query.Query, number string) (exist bool, err error) {
	q := tx.UserBasic
	_, err = q.WithContext(ctx).Select(q.ID).Where(q.Number.Eq(number)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo *GenRepo) AddSignInLog(ctx context.Context, tx *query.Query, req *userv1.SignInReq, accessClaims, refreshClaims *JwtClaims) (err error) {
	return tx.UserSignLog.WithContext(ctx).Create(&model.UserSignLog{
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
func (repo *GenRepo) GetActiveSignInLogs(ctx context.Context, tx *query.Query, uid int64) (logs []*model.UserSignLog, err error) {
	q := tx.UserSignLog
	now := time.Now().UnixMilli()
	logs, err = q.WithContext(ctx).Where(
		q.UID.Eq(int64(uid)),
		q.Status.Eq(int16(enumsv1.SessionStatus_SESSION_STATUS_OK)),
		q.RefreshExpiredAt.Gt(now),
	).Order(q.RefreshExpiredAt.Desc()).Find()
	return
}

func (repo *GenRepo) GetSignInLogByAccess(ctx context.Context, tx *query.Query, accessSessionID string) (log *model.UserSignLog, err error) {
	q := tx.UserSignLog
	log, err = q.WithContext(ctx).Where(q.AccessSessionID.Eq(accessSessionID)).Take()
	return
}

func (repo *GenRepo) GetSignInLogByRefresh(ctx context.Context, tx *query.Query, refreshSessionID string) (log *model.UserSignLog, err error) {
	q := tx.UserSignLog
	log, err = q.WithContext(ctx).Where(q.RefreshSessionID.Eq(refreshSessionID)).Take()
	return
}

func (repo *GenRepo) UpdateSignInLogByID(ctx context.Context, tx *query.Query, log *model.UserSignLog) (err error) {
	q := tx.UserSignLog
	_, err = q.WithContext(ctx).Where(q.ID.Eq(log.ID)).Updates(log)
	return
}

func (repo *GenRepo) UpdateSignInLogsStatus(ctx context.Context, tx *query.Query, ids []int64, status enumsv1.SessionStatus) (err error) {
	q := tx.UserSignLog
	_, err = q.WithContext(ctx).Where(q.ID.In(ids...)).Update(q.Status, int16(status))
	return
}

func (repo *GenRepo) UpdateUserBasicByUid(ctx context.Context, tx *query.Query, basic *model.UserBasic) (err error) {
	q := tx.UserBasic
	_, err = q.WithContext(ctx).Where(q.ID.Eq(basic.ID)).Updates(basic)
	return
}

func (repo *GenRepo) PagingGetSignInLogs(ctx context.Context, tx *query.Query, req *userv1.PagingGetSignInLogsReq) (resp *orm.PageResult[model.UserSignLog], err error) {
	q := tx.UserSignLog
	db := q.WithContext(ctx).Where(q.Biz.Eq(req.Biz))
	if req.Asc {
		db = db.Order(q.ID.Asc())
	} else {
		db = db.Order(q.ID.Desc())
	}
	if req.Uid != nil {
		db = db.Where(q.UID.Eq(*req.Uid))
	}
	if len(req.Statuses) > 0 {
		var statuses = make([]int16, len(req.Statuses))
		for i, status := range req.Statuses {
			statuses[i] = int16(status)
		}
		statuses = slicex.Unique(statuses)
		if len(statuses) == 1 {
			db = db.Where(q.Status.Eq(statuses[0]))
		} else {
			db = db.Where(q.Status.In(statuses...))
		}
	}
	if len(req.Types) > 0 {
		var types = make([]int16, len(req.Types))
		for i, typ := range req.Types {
			types[i] = int16(typ)
		}
		types = slicex.Unique(types)
		if len(types) == 1 {
			db = db.Where(q.Type.Eq(types[0]))
		} else {
			db = db.Where(q.Type.In(types...))
		}
	}
	return orm.Paginate[model.UserSignLog](db.UnderlyingDB(), uint32(req.Page), uint32(req.Size))
}

func (repo *GenRepo) PagingGetUsers(ctx context.Context, tx *query.Query, req *userv1.PagingGetUsersReq) (resp *orm.PageResult[model.UserBasic], err error) {
	q := tx.UserBasic
	db := q.WithContext(ctx).Where(q.Biz.Eq(req.Biz))
	if req.ShowDeleted {
		db = db.Unscoped()
	}
	if req.Asc {
		db = db.Order(q.ID.Asc())
	} else {
		db = db.Order(q.ID.Desc())
	}
	if req.Uid != nil {
		db = db.Where(q.ID.Eq(*req.Uid))
	}
	if req.Number != nil {
		db = db.Where(q.Number.Eq(*req.Number))
	}
	if req.Name != nil {
		db = db.Where(q.Name.Like(*req.Name + "%"))
	}
	if req.Gender != nil {
		db = db.Where(q.Gender.Eq(int16(*req.Gender)))
	}
	if req.Phone != nil {
		db = db.Where(
			q.PhoneCountryCode.Eq(req.Phone.CountryCode),
			q.Phone.Eq(req.Phone.Number),
		)
	}
	if req.Email != nil {
		db = db.Where(q.Email.Eq(*req.Email))
	}
	if req.CountryCode != nil {
		db = db.Where(q.CountryCode.Eq(*req.CountryCode))
		if req.ProvinceCode != nil {
			db = db.Where(q.ProvinceCode.Eq(*req.ProvinceCode))
			if req.CityCode != nil {
				db = db.Where(q.CityCode.Eq(*req.CityCode))
				if req.DistrictCode != nil {
					db = db.Where(q.DistrictCode.Eq(*req.DistrictCode))
				}
			}
		}
	}
	if req.SignUpType != nil {
		db = db.Where(q.SignupType.Eq(int16(*req.SignUpType)))
	}
	if req.Status != nil {
		db = db.Where(q.Status.Eq(int16(*req.Status)))
	}
	if len(req.Source) > 0 {
		var sources = make([]int16, len(req.Source))
		for i, source := range req.Source {
			sources[i] = int16(source)
		}
		sources = slicex.Unique(sources)
		if len(sources) == 1 {
			db = db.Where(q.Source.Eq(sources[0]))
		} else {
			db = db.Where(q.Source.In(sources...))
		}
	}
	if req.Type != nil {
		db = db.Where(q.Type.Eq(int16(*req.Type)))
	}
	if req.Level != nil {
		db = db.Where(q.Level.Eq(*req.Level))
	}
	if req.PhoneVerified != nil {
		if *req.PhoneVerified {
			db = db.Where(q.PhoneVerified.Eq(int16(enumsv1.Verified_VERIFIED_VERIFIED)))
		} else {
			db = db.Where(q.PhoneVerified.Eq(int16(enumsv1.Verified_VERIFIED_UNVERIFIED)))
		}
	}
	if req.EmailVerified != nil {
		if *req.EmailVerified {
			db = db.Where(q.EmailVerified.Eq(int16(enumsv1.Verified_VERIFIED_VERIFIED)))
		} else {
			db = db.Where(q.EmailVerified.Eq(int16(enumsv1.Verified_VERIFIED_UNVERIFIED)))
		}
	}
	if req.BirthdayStart != nil && req.BirthdayEnd != nil {
		start := time.Date(int(req.BirthdayStart.Year), time.Month(req.BirthdayStart.Month), int(req.BirthdayStart.Day), 0, 0, 0, 0, time.UTC)
		end := time.Date(int(req.BirthdayEnd.Year), time.Month(req.BirthdayEnd.Month), int(req.BirthdayEnd.Day), 0, 0, 0, 0, time.UTC)
		db = db.Where(q.Birthday.Between(start, end))
	}
	if req.CreatedAtStart != nil && req.CreatedAtEnd != nil {
		db = db.Where(q.CreatedAt.Between(req.CreatedAtStart.AsTime().UnixMilli(), req.CreatedAtEnd.AsTime().UnixMilli()))
	}
	if req.UpdatedAtStart != nil && req.UpdatedAtEnd != nil {
		db = db.Where(q.UpdatedAt.Between(req.UpdatedAtStart.AsTime().UnixMilli(), req.UpdatedAtEnd.AsTime().UnixMilli()))
	}
	if req.DeletedAtStart != nil && req.DeletedAtEnd != nil {
		db = db.Where(q.DeletedAt.Between(req.DeletedAtStart.AsTime().UnixMilli(), req.DeletedAtEnd.AsTime().UnixMilli()))
	}
	return orm.Paginate[model.UserBasic](db.UnderlyingDB(), uint32(req.Page), uint32(req.Size))
}

func (repo *GenRepo) DeleteUserBasic(ctx context.Context, tx *query.Query, uid int64) error {
	q := tx.UserBasic
	_, err := q.WithContext(ctx).Where(q.ID.Eq(uid)).Delete()
	return err
}

func (repo *GenRepo) DeleteUsersBasic(ctx context.Context, tx *query.Query, uids []int64) error {
	q := tx.UserBasic
	_, err := q.WithContext(ctx).Where(q.ID.In(uids...)).Delete()
	return err
}

func (repo *GenRepo) DeleteUserAuth(ctx context.Context, tx *query.Query, uid int64) error {
	q := tx.UserAuth
	_, err := q.WithContext(ctx).Where(q.ID.Eq(uid)).Delete()
	return err
}

func (repo *GenRepo) DeleteUsersAuth(ctx context.Context, tx *query.Query, uids []int64) error {
	q := tx.UserAuth
	_, err := q.WithContext(ctx).Where(q.ID.In(uids...)).Delete()
	return err
}

func (repo *GenRepo) DeleteUserSignLogs(ctx context.Context, tx *query.Query, uid int64) error {
	q := tx.UserSignLog
	_, err := q.WithContext(ctx).Where(q.ID.Eq(uid)).Delete()
	return err
}

func (repo *GenRepo) DeleteUsersSignLogs(ctx context.Context, tx *query.Query, uids []int64) error {
	q := tx.UserSignLog
	_, err := q.WithContext(ctx).Where(q.ID.In(uids...)).Delete()
	return err
}

func (repo *GenRepo) GetUserAuthByUid(ctx context.Context, tx *query.Query, uid int64) (auth []*model.UserAuth, err error) {
	q := tx.UserAuth
	return q.WithContext(ctx).Where(q.UID.Eq(uid)).Find()
}

func (repo *GenRepo) UnbindUserAuth(ctx context.Context, tx *query.Query, uid int64, id string) error {
	q := tx.UserAuth
	_, err := q.WithContext(ctx).Where(
		q.ID.Eq(uid),
		q.Identifier.Eq(id),
	).Delete()
	return err
}

func (repo *GenRepo) getIdentifier(req *userv1.SignInReq) (identifier string) {
	if req.AuthType == enumsv1.AuthType_AUTH_TYPE_PHONE_CAPTCHA || req.AuthType == enumsv1.AuthType_AUTH_TYPE_PHONE_PASSWORD {
		return req.PhoneNumber.CountryCode + req.PhoneNumber.Number
	}
	return req.Identifier
}
