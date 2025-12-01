package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	typesv1 "github.com/byteflowing/proto/gen/go/types/v1"
	"github.com/golang-jwt/jwt/v5"
)

func IsUserValid(st int16) bool {
	if st != int16(enumsv1.UserStatus_USER_STATUS_DISABLED) {
		return false
	}
	return true
}

func LocationToString(location *typesv1.Location) *string {
	if location == nil {
		return nil
	}
	lnglat := fmt.Sprintf(locationFormat, location.Lng, location.Lat)
	return &lnglat
}

func ParseLocationFromString(location *string) *typesv1.Location {
	if location == nil {
		return nil
	}
	lnglat := strings.Split(*location, ",")
	lng, _ := strconv.ParseFloat(lnglat[0], 64)
	lat, _ := strconv.ParseFloat(lnglat[1], 64)
	return &typesv1.Location{
		Lat: lat,
		Lng: lng,
	}
}

func GetJwtJti(claims jwt.MapClaims) string {
	jti, ok := claims[JwtJtiKey].(string)
	if !ok {
		return ""
	}
	return jti
}

func GetJwtExp(claims jwt.MapClaims) *time.Time {
	exp, err := claims.GetExpirationTime()
	if err != nil {
		return nil
	}
	expiredAt := exp.Time
	return &expiredAt
}

func GetTokenType(claims jwt.MapClaims) enumsv1.TokenType {
	t, _ := claims[JwtTokenTypeKey].(string)
	return enumsv1.TokenType(enumsv1.TokenType_value[t])
}

func GetTokenTenantID(claims jwt.MapClaims) string {
	tenant, _ := claims[JwtTenantIDKey].(string)
	return tenant
}

func GetTokenNumber(claims jwt.MapClaims) string {
	number, _ := claims[JwtNumberKey].(string)
	return number
}

func GetTokenUserType(claims jwt.MapClaims) *int32 {
	userType, ok := claims[JwtTypeKey].(string)
	if !ok {
		return nil
	}
	t, err := strconv.ParseInt(userType, 10, 32)
	if err != nil {
		return nil
	}
	i32t := int32(t)
	return &i32t
}

func GetTokenUserLevel(claims jwt.MapClaims) *int32 {
	userLevel, ok := claims[JwtLevelKey].(string)
	if !ok {
		return nil
	}
	l, err := strconv.ParseInt(userLevel, 10, 32)
	if err != nil {
		return nil
	}
	i32l := int32(l)
	return &i32l
}
