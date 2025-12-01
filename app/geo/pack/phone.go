package pack

import (
	"github.com/byteflowing/base/app/geo/dal/model"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func PhoneCodeProtoToModel(phoneCodeProto *geov1.PhoneCode) *model.GeoPhoneCode {
	v := &model.GeoPhoneCode{
		Name:      phoneCodeProto.GetName(),
		PhoneCode: phoneCodeProto.GetPhoneCode(),
		IsActive:  phoneCodeProto.IsActive,
	}
	v.MultiLang = MultiLangToJsonString(phoneCodeProto.MultiLang)
	return v
}

// PhoneCodeModelToProto 将数据库模型转换为GeoPhoneCode protobuf消息
func PhoneCodeModelToProto(phoneCode *model.GeoPhoneCode) *geov1.GeoPhoneCode {
	v := &geov1.GeoPhoneCode{
		Id:        phoneCode.ID,
		Name:      phoneCode.Name,
		PhoneCode: phoneCode.PhoneCode,
	}
	if phoneCode.CreatedAt != nil {
		v.CreatedAt = timestamppb.New(*phoneCode.CreatedAt)
	}
	if phoneCode.UpdatedAt != nil {
		v.UpdatedAt = timestamppb.New(*phoneCode.UpdatedAt)
	}
	if phoneCode.DeletedAt.Valid {
		v.DeletedAt = timestamppb.New(phoneCode.DeletedAt.Time)
	}
	v.MultiLang = MultiLangFromJsonString(phoneCode.MultiLang)
	return v
}
