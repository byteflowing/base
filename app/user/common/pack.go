package common

import (
	"github.com/byteflowing/base/app/user/dal/model"
	"github.com/byteflowing/base/pkg/utils/trans"
	"github.com/byteflowing/proto/gen/go/enums/v1"
	typesv1 "github.com/byteflowing/proto/gen/go/types/v1"
	userv1 "github.com/byteflowing/proto/gen/go/user/v1"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserModelToUser(m *model.UserAccount) *userv1.User {
	u := &userv1.User{
		Uid:         m.ID,
		TenantId:    m.TenantID,
		Number:      m.Number,
		Name:        trans.Deref(m.Name),
		Alias:       trans.Deref(m.Alias_),
		Avatar:      trans.Deref(m.Avatar),
		Email:       m.Email,
		Addr:        trans.Deref(m.Addr),
		Status:      enumsv1.UserStatus(m.Status),
		Source:      enumsv1.UserSource(m.Source),
		SignUpType:  enumsv1.SignInType(m.SignupType),
		UserType:    int32(trans.Deref(m.Type)),
		UserLevel:   int32(trans.Deref(m.Level)),
		Ext:         trans.Deref(m.Ext),
		PhoneVerify: m.PhoneVerified,
		EmailVerify: m.EmailVerified,
	}
	if m.Birthday != nil {
		u.Birthday = &date.Date{
			Year:  int32(m.Birthday.Year()),
			Month: int32(m.Birthday.Month()),
			Day:   int32(m.Birthday.Day()),
		}
	}
	if m.PhoneCountryCode != "" && m.Phone != "" {
		u.PhoneNumber = &typesv1.PhoneNumber{
			CountryCode: m.PhoneCountryCode,
			Number:      m.Phone,
		}
	}
	if m.CountryCode != "" && m.ProvinceCode != "" && m.CityCode != "" && m.DistrictCode != "" {
		u.Region = &typesv1.AdminRegion{
			CountryCode:  m.CountryCode,
			ProvinceCode: m.ProvinceCode,
			CityCode:     m.CityCode,
			DistrictCode: m.DistrictCode,
		}
	}
	if m.Gender != nil {
		u.Gender = enumsv1.Gender(*m.Gender)
	}
	u.RegisterAgent = &userv1.Agent{
		Ip:       m.RegisterIP,
		Agent:    m.RegisterAgent,
		Device:   m.RegisterDevice,
		Location: ParseLocationFromString(m.RegisterLocation),
	}
	if m.PasswordUpdatedAt != nil {
		u.PasswordUpdatedAt = timestamppb.New(*m.PasswordUpdatedAt)
	}
	if m.DeletedAt.Valid {
		u.DeleteAt = timestamppb.New(m.DeletedAt.Time)
	}
	if m.UpdatedAt != nil {
		u.UpdatedAt = timestamppb.New(*m.UpdatedAt)
	}
	if m.CreatedAt != nil {
		u.CreatedAt = timestamppb.New(*m.CreatedAt)
	}
	return u
}

func ClaimsToJwtClaims(claims jwt.MapClaims, extraKey []string) *userv1.JwtClaims {
	iss, _ := claims.GetIssuer()
	sub, _ := claims.GetSubject()
	iat, _ := claims.GetIssuedAt()
	nbf, _ := claims.GetNotBefore()
	exp, _ := claims.GetExpirationTime()
	c := &userv1.JwtClaims{
		Iss:       iss,
		Sub:       sub,
		Iat:       iat.Unix(),
		Nbf:       nbf.Unix(),
		Exp:       exp.Unix(),
		Jti:       GetJwtJti(claims),
		TokenType: GetTokenType(claims),
		TenantId:  GetTokenTenantID(claims),
		Number:    GetTokenNumber(claims),
		Type:      GetTokenUserType(claims),
		Level:     GetTokenUserLevel(claims),
	}
	extra := make(map[string]string, len(extraKey))
	for _, k := range extraKey {
		v, ok := claims[k]
		if !ok {
			continue
		}
		extra[k] = v.(string)
	}
	c.Extra = extra
	return c
}
