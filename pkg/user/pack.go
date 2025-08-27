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

func userBasicToUserInfo(userBasic *model.UserBasic) *userv1.UserInfo {
	info := &userv1.UserInfo{
		Uid:    userBasic.ID,
		Number: userBasic.Number,
		Status: enumsv1.UserStatus(userBasic.Status),
		Name:   userBasic.Name,
		Alias:  userBasic.Alias_,
		Avatar: userBasic.Avatar,
		Addr:   userBasic.Addr,
		Ext:    userBasic.Ext,
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
