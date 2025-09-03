package user

import (
	"time"

	"github.com/byteflowing/base/dal/model"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/go-common/crypto"
	"github.com/byteflowing/go-common/trans"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func signUpReqToUserBasic(
	req *userv1.SignUpReq,
	globalId common.GlobalIdGenerator,
	shortId *common.ShortIDGenerator,
	passHasher *crypto.PasswordHasher,
) (userBasic *model.UserBasic, err error) {
	uid, err := globalId.GetID()
	if err != nil {
		return nil, err
	}
	userBasic = &model.UserBasic{
		Biz:            req.Biz,
		ID:             uid,
		Name:           req.Name,
		Alias_:         req.Alias,
		Avatar:         req.Avatar,
		Addr:           req.Addr,
		Level:          req.Level,
		SignupType:     int16(req.AuthType),
		RegisterIP:     req.Ip,
		RegisterDevice: req.Device,
		RegisterAgent:  req.Agent,
		Ext:            req.Ext,
	}
	if req.Number == nil || len(*req.Number) == 0 {
		number, err := shortId.GetID()
		if err != nil {
			return nil, err
		}
		userBasic.Number = number
	}
	if req.Password != nil {
		passwd, err := passHasher.HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		userBasic.Password = trans.Ref(passwd)
		userBasic.PasswordUpdatedAt = time.Now().UnixMilli()
	}
	if req.Gender != nil {
		userBasic.Gender = trans.Ref(int16(*req.Gender))
	}
	if req.Birthday != nil {
		birthday := time.Date(int(req.Birthday.Year), time.Month(req.Birthday.Month), int(req.Birthday.Day), 0, 0, 0, 0, time.UTC)
		userBasic.Birthday = trans.Ref(birthday)
	}
	if req.Region != nil {
		userBasic.CountryCode = req.Region.CountryCode
		userBasic.ProvinceCode = req.Region.ProvinceCode
		userBasic.CityCode = req.Region.CityCode
		userBasic.DistrictCode = req.Region.DistrictCode
	}
	if req.Status != nil {
		userBasic.Status = int16(*req.Status)
	} else {
		userBasic.Status = int16(enumsv1.UserStatus_USER_STATUS_OK)
	}
	if req.Source != nil {
		userBasic.Source = int16(*req.Source)
	}
	switch req.AuthType {
	case enumsv1.AuthType_AUTH_TYPE_EMAIL_CAPTCHA:
		userBasic.Email = *req.Email
		userBasic.EmailVerified = int16(enumsv1.Verified_VERIFIED_VERIFIED)
	case enumsv1.AuthType_AUTH_TYPE_PHONE_CAPTCHA:
		userBasic.PhoneCountryCode = req.Phone.CountryCode
		userBasic.Phone = req.Phone.Number
		userBasic.PhoneVerified = int16(enumsv1.Verified_VERIFIED_VERIFIED)
	}
	return userBasic, nil
}

func createUserReqToUserBasic(
	req *userv1.CreateUserReq,
	globalId common.GlobalIdGenerator,
	shortId *common.ShortIDGenerator,
	passHasher *crypto.PasswordHasher,
) (basic *model.UserBasic, err error) {
	uid, err := globalId.GetID()
	if err != nil {
		return nil, err
	}
	basic = &model.UserBasic{
		ID:            uid,
		Biz:           req.Biz,
		Name:          req.Name,
		Alias_:        req.Alias,
		Avatar:        req.Avatar,
		Email:         trans.StringValue(req.Email),
		Addr:          req.Addr,
		PhoneVerified: int16(enumsv1.Verified_VERIFIED_UNVERIFIED),
		EmailVerified: int16(enumsv1.Verified_VERIFIED_UNVERIFIED),
		Level:         req.Level,
		Ext:           req.Ext,
	}
	if req.Number == nil || len(*req.Number) == 0 {
		number, err := shortId.GetID()
		if err != nil {
			return nil, err
		}
		basic.Number = number
	}
	if req.Password != nil {
		passwd, err := passHasher.HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		basic.Password = trans.Ref(passwd)
		basic.PasswordUpdatedAt = time.Now().UnixMilli()
	}
	if req.Gender != nil {
		basic.Gender = trans.Int16(int16(*req.Gender))
	}
	if req.Birthday != nil {
		basic.Birthday = trans.Ref(time.Date(int(req.Birthday.Year), time.Month(req.Birthday.Month), int(req.Birthday.Day), 0, 0, 0, 0, time.UTC))
	}
	if req.Phone != nil {
		basic.Phone = req.Phone.Number
		basic.PhoneCountryCode = req.Phone.CountryCode
	}
	if req.Region != nil {
		basic.CountryCode = req.Region.CountryCode
		basic.ProvinceCode = req.Region.ProvinceCode
		basic.CityCode = req.Region.CityCode
		basic.DistrictCode = req.Region.DistrictCode
	}
	if req.Status != nil {
		basic.Status = int16(*req.Status)
	}
	if req.Source != nil {
		basic.Source = int16(*req.Source)
	}
	if req.Type != nil {
		basic.Type = trans.Int16(int16(*req.Type))
	}
	return basic, nil
}

func updateUserReqToUserBasic(
	req *userv1.UpdateUserReq,
	passHasher *crypto.PasswordHasher,
) (userBasic *model.UserBasic, err error) {
	userBasic = &model.UserBasic{
		ID:     req.Uid,
		Name:   req.Name,
		Alias_: req.Alias,
		Avatar: req.Avatar,
		Addr:   req.Addr,
		Level:  req.Level,
		Ext:    req.Ext,
	}
	if req.Number != nil {
		userBasic.Number = *req.Number
	}
	if req.Password != nil {
		passwd, err := passHasher.HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		userBasic.Password = trans.Ref(passwd)
		userBasic.PasswordUpdatedAt = time.Now().UnixMilli()
	}
	if req.Gender != nil {
		userBasic.Gender = trans.Int16(int16(*req.Gender))
	}
	if req.Birthday != nil {
		userBasic.Birthday = trans.Ref(time.Date(int(req.Birthday.Year), time.Month(req.Birthday.Month), int(req.Birthday.Day), 0, 0, 0, 0, time.UTC))
	}
	if req.Phone != nil {
		userBasic.Phone = req.Phone.Number
		userBasic.PhoneCountryCode = req.Phone.CountryCode
	}
	if req.Email != nil {
		userBasic.Email = *req.Email
	}
	if req.Region != nil {
		userBasic.CountryCode = req.Region.CountryCode
		userBasic.ProvinceCode = req.Region.ProvinceCode
		userBasic.CityCode = req.Region.CityCode
		userBasic.DistrictCode = req.Region.DistrictCode
	}
	if req.Status != nil {
		userBasic.Status = int16(*req.Status)
	}
	if req.Type != nil {
		userBasic.Type = trans.Int16(int16(*req.Type))
	}
	return userBasic, nil
}

func userBasicToUserInfo(userBasic *model.UserBasic) (basic *userv1.UserInfo) {
	info := &userv1.UserInfo{
		Uid:               userBasic.ID,
		Biz:               userBasic.Biz,
		Number:            userBasic.Number,
		Status:            enumsv1.UserStatus(userBasic.Status),
		Source:            enumsv1.UserSource(userBasic.Source),
		SignUpType:        enumsv1.AuthType(userBasic.SignupType),
		UserLevel:         userBasic.Level,
		Name:              userBasic.Name,
		Alias:             userBasic.Alias_,
		Avatar:            userBasic.Avatar,
		Addr:              userBasic.Addr,
		Ext:               userBasic.Ext,
		PhoneVerify:       verifyToBool(userBasic.PhoneVerified),
		EmailVerify:       verifyToBool(userBasic.EmailVerified),
		RegisterIp:        trans.StringValue(userBasic.RegisterIP),
		RegisterDevice:    trans.StringValue(userBasic.RegisterDevice),
		RegisterAgent:     trans.StringValue(userBasic.RegisterAgent),
		PasswordUpdatedAt: timestamppb.New(time.UnixMilli(userBasic.PasswordUpdatedAt)),
		UpdatedAt:         timestamppb.New(time.UnixMilli(userBasic.UpdatedAt)),
		CreatedAt:         timestamppb.New(time.UnixMilli(userBasic.CreatedAt)),
	}
	if userBasic.DeletedAt != nil {
		info.DeleteAt = timestamppb.New(time.UnixMilli(*userBasic.DeletedAt))
	}
	if userBasic.Type != nil {
		info.UserType = trans.Int32(int32(*userBasic.Type))
	}
	if userBasic.Gender != nil {
		info.Gender = trans.Ref(enumsv1.Gender(*userBasic.Gender))
	}
	if userBasic.Birthday != nil {
		info.Birthday = &date.Date{
			Year:  int32(userBasic.Birthday.Year()),
			Month: int32(userBasic.Birthday.Month()),
			Day:   int32(userBasic.Birthday.Day()),
		}
	}
	if userBasic.Phone != "" {
		info.PhoneNumber = &commonv1.PhoneNumber{
			CountryCode: userBasic.PhoneCountryCode,
			Number:      userBasic.Phone,
		}
	}
	if userBasic.Email != "" {
		info.Email = trans.Ref(userBasic.Email)
	}
	if userBasic.CountryCode != "" {
		info.Region = &commonv1.AdminRegion{
			CountryCode:  userBasic.CountryCode,
			ProvinceCode: userBasic.ProvinceCode,
			CityCode:     userBasic.CityCode,
			DistrictCode: userBasic.DistrictCode,
		}
	}
	return info
}

func userSignInModelToSignInLog(model *model.UserSignLog) (log *userv1.SignInLog) {
	log = &userv1.SignInLog{
		Id:               model.ID,
		Uid:              model.UID,
		Type:             enumsv1.AuthType(model.Type),
		Status:           enumsv1.SessionStatus(model.Status),
		AccessTokenId:    model.AccessSessionID,
		RefreshTokenId:   model.RefreshSessionID,
		Ip:               model.IP,
		Location:         model.Location,
		Agent:            model.Agent,
		Device:           model.Device,
		DeletedAt:        model.DeletedAt,
		AccessExpiredAt:  timestamppb.New(time.Unix(model.AccessExpiredAt, 0)),
		RefreshExpiredAt: timestamppb.New(time.Unix(model.RefreshExpiredAt, 0)),
		UpdatedAt:        timestamppb.New(time.UnixMilli(model.UpdatedAt)),
		CreatedAt:        timestamppb.New(time.UnixMilli(model.CreatedAt)),
	}
	return
}

func userAuthModelToAuth(m *model.UserAuth) (auth *userv1.UserAuth) {
	return &userv1.UserAuth{
		Id:         m.ID,
		Uid:        m.UID,
		Type:       enumsv1.AuthType(m.Type),
		Status:     enumsv1.AuthStatus(m.Status),
		AppId:      m.Appid,
		Identifier: m.Identifier,
		Biz:        m.Biz,
	}
}

func verifyToBool(v int16) bool {
	verify := enumsv1.Verified(v)
	if verify == enumsv1.Verified_VERIFIED_VERIFIED {
		return true
	}
	return false
}
