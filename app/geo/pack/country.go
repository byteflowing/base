package pack

import (
	"github.com/byteflowing/base/app/geo/dal/model"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CountryCodeProtoToModels(p *geov1.CountryCode) *model.GeoCountry {
	v := &model.GeoCountry{
		Cca2:         p.GetCca2(),
		Cca3:         p.GetCca3(),
		Ccn3:         p.GetCcn3(),
		Flag:         p.GetFlag(),
		Continent:    p.GetContinent(),
		SubContinent: p.GetSubContinent(),
		Independent:  p.GetIndependent(),
		IsActive:     p.IsActive,
	}
	v.MultiLang = MultiLangToJsonString(p.MultiLang)
	return v
}

// CountryModelToGeoProto 将数据库模型转换为GeoCountry protobuf消息
func CountryModelToGeoProto(country *model.GeoCountry) *geov1.GeoCountry {
	v := &geov1.GeoCountry{
		Id:           country.ID,
		Cca2:         country.Cca2,
		Cca3:         country.Cca3,
		Ccn3:         country.Ccn3,
		Flag:         country.Flag,
		Continent:    country.Continent,
		SubContinent: country.SubContinent,
		Independent:  country.Independent,
	}
	v.MultiLang = MultiLangFromJsonString(country.MultiLang)
	if country.CreatedAt != nil {
		v.CreatedAt = timestamppb.New(*country.CreatedAt)
	}
	if country.UpdatedAt != nil {
		v.UpdatedAt = timestamppb.New(*country.UpdatedAt)
	}
	if country.DeletedAt.Valid {
		v.DeletedAt = timestamppb.New(country.DeletedAt.Time)
	}
	return v
}
