package migrate

import (
	"context"
	"fmt"

	"github.com/byteflowing/base/app/geo/dal/model"
	"github.com/byteflowing/base/app/geo/pack"
	"github.com/byteflowing/base/pkg/logx"
	"github.com/byteflowing/base/pkg/sdk/tencent/lbs"
	"github.com/byteflowing/base/pkg/utils/trans"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
)

func (m *Migrate) MigrateRegions() error {
	logx.Warn("-------------------------importing region codes started-------------------------------------------")
	defer logx.Warn("-------------------------importing region codes ended---------------------------------------")
	ctx := context.Background()
	for _, cfg := range m.cfg.RegionSource {
		switch cfg.Cca2 {
		case "CN":
			if err := m.migrateRegionsFromTencentMap(ctx, cfg); err != nil {
				return err
			}
		default:
			panic(fmt.Sprintf("unsupported cca type: %s", cfg.Cca2))
		}
	}
	return nil
}

func (m *Migrate) migrateRegionsFromTencentMap(ctx context.Context, cfg *geov1.Config_RegionSource) (err error) {
	msgStart := fmt.Sprintf("-----------------------------[country:%s]importing region codes started---------------------------------------", cfg.Cca2)
	msgEnd := fmt.Sprintf("-----------------------------[country:%s]importing region codes ended---------------------------------------", cfg.Cca2)
	logx.Info(msgStart)
	defer logx.Info(msgEnd)
	q := m._query.GeoRegion
	if _, err := q.WithContext(ctx).Where(q.ID.Gt(0), q.CountryCca2.Eq(cfg.Cca2)).Take(); err == nil {
		logx.Warn("region codes already exist, so ignore import")
		return
	}
	mapService := lbs.NewMapService()
	resp, err := mapService.GetDistricts(ctx, &mapsv1.TencentGetDistrictsReq{Key: cfg.MapKey})
	if err != nil {
		return err
	}
	if resp == nil || len(resp.Result) == 0 {
		return fmt.Errorf("region code not found")
	}
	regions := m.convertTencentDistrictsToRegionCodes(resp.Result, cfg.Source, cfg.Cca2)
	var models = make([]*model.GeoRegion, 0, len(regions))
	for _, region := range regions {
		models = append(models, pack.RegionCodeProtoToModels(region))
	}
	if len(models) == 0 {
		return fmt.Errorf("region code not found")
	}
	return q.WithContext(ctx).Create(models...)
}

func (m *Migrate) convertTencentDistrictsToRegionCodes(
	districts []*mapsv1.TencentDistrict,
	source enumsv1.RegionSource,
	countryCca2 string,
) []*geov1.RegionCode {
	var regionCodes []*geov1.RegionCode
	var walk func(nodes []*mapsv1.TencentDistrict, parentCode *string, level int32)
	walk = func(nodes []*mapsv1.TencentDistrict, parentCode *string, level int32) {
		for _, node := range nodes {
			if node == nil {
				continue
			}
			lang := make(map[string]string)
			lang[enumsv1.Language_LANGUAGE_ZH_CN.String()] = node.Fullname
			rc := &geov1.RegionCode{
				CountryCca2: trans.Ref(countryCca2),
				Source:      trans.Ref(source),
				Level:       trans.Ref(enumsv1.RegionLevel(level)),
				ParentCode:  parentCode,
				Code:        trans.Ref(node.Id),
				IsActive:    trans.Ref(true),
				MultiLang:   lang,
			}
			regionCodes = append(regionCodes, rc)
			if len(node.Districts) > 0 {
				walk(node.Districts, &node.Id, level+1)
			}
		}
	}
	walk(districts, nil, 1)
	return regionCodes
}
