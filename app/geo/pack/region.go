package pack

import (
	"github.com/byteflowing/base/app/geo/dal/model"
	"github.com/byteflowing/base/pkg/utils/trans"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func RegionCodeProtoToModels(req *geov1.RegionCode) *model.GeoRegion {
	v := &model.GeoRegion{
		CountryCca2: req.GetCountryCca2(),
		Source:      int16(req.GetSource()),
		ParentCode:  req.GetParentCode(),
		Code:        req.GetCode(),
		Level:       int16(req.GetLevel()),
		IsActive:    req.IsActive,
	}
	v.MultiLang = MultiLangToJsonString(req.MultiLang)
	return v
}

// RegionModelToProto 将数据库模型转换为GeoRegion protobuf消息
func RegionModelToProto(region *model.GeoRegion) *geov1.GeoRegion {
	v := &geov1.GeoRegion{
		Id:         region.ID,
		Source:     enumsv1.RegionSource(region.Source),
		ParentCode: region.ParentCode,
		Code:       region.Code,
		Level:      enumsv1.RegionLevel(region.Level),
		IsActive:   trans.Deref(region.IsActive),
	}
	v.MultiLang = MultiLangFromJsonString(region.MultiLang)
	if region.CreatedAt != nil {
		v.CreatedAt = timestamppb.New(*region.CreatedAt)
	}
	if region.UpdatedAt != nil {
		v.UpdatedAt = timestamppb.New(*region.UpdatedAt)
	}
	if region.DeletedAt.Valid {
		v.DeletedAt = timestamppb.New(region.DeletedAt.Time)
	}
	return v
}
