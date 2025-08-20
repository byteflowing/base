package user

import (
	"github.com/byteflowing/base/dal/model"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
)

func isDisabled(userBasic *model.UserBasic) bool {
	return userBasic.Status == int16(enumsv1.UserStatus_USER_STATUS_DISABLED)
}
