package service

import (
	"context"
	"errors"

	"github.com/byteflowing/base/app/geo/cache"
	"github.com/byteflowing/base/app/geo/dal/model"
	"github.com/byteflowing/base/app/geo/dal/query"
	"github.com/byteflowing/base/app/geo/pack"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/db"
	"github.com/byteflowing/base/pkg/utils/trans"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
)

func (g *GeoService) AddGeoRegion(ctx context.Context, req *geov1.AddGeoRegionReq) (*geov1.AddGeoRegionResp, error) {
	results, err := g.addRegions(ctx, []*geov1.RegionCode{req.GetRegionCode()})
	if err != nil {
		return nil, err
	}
	var result *geov1.AddRegionResult
	if len(results) > 0 {
		result = results[0]
	}
	return &geov1.AddGeoRegionResp{Result: result}, nil
}

func (g *GeoService) AddGeoRegions(ctx context.Context, req *geov1.AddGeoRegionsReq) (*geov1.AddGeoRegionsResp, error) {
	results, err := g.addRegions(ctx, req.GetRegionCodes())
	if err != nil {
		return nil, err
	}
	return &geov1.AddGeoRegionsResp{Results: results}, nil
}

func (g *GeoService) UpdateGeoRegion(ctx context.Context, req *geov1.UpdateGeoRegionReq) (*geov1.UpdateGeoRegionResp, error) {
	cca2, err := g.getCca2ByRegionID(ctx, g.db, req.GetRegionId())
	if err != nil {
		return nil, err
	}
	key := g.cache.GetCountryCca2RegionKey(cca2)
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	param := pack.RegionCodeProtoToModels(req.GetRegionCode())
	err = g.db.Transaction(func(tx *query.Query) error {
		q := tx.GeoRegion
		if _, err = q.WithContext(ctx).Where(q.ID.Eq(req.RegionId)).Updates(param); err != nil {
			return err
		}
		return g.cache.DeleteRegionsByCountryCca2(ctx, cca2)
	})
	if err != nil {
		return nil, err
	}
	return &geov1.UpdateGeoRegionResp{}, err
}

func (g *GeoService) DeleteGeoRegion(ctx context.Context, req *geov1.DeleteGeoRegionReq) (*geov1.DeleteGeoRegionResp, error) {
	cca2, err := g.getCca2ByRegionID(ctx, g.db, req.RegionId)
	if err != nil {
		return nil, err
	}
	key := g.cache.GetCountryCca2RegionKey(cca2)
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	err = g.db.Transaction(func(tx *query.Query) error {
		if _, err := tx.GeoRegion.WithContext(ctx).Where(tx.GeoRegion.ID.Eq(req.RegionId)).Delete(); err != nil {
			return err
		}
		return g.cache.DeleteRegionsByCountryCca2(ctx, cca2)
	})
	if err != nil {
		return nil, err
	}
	return &geov1.DeleteGeoRegionResp{}, nil
}

func (g *GeoService) GetGeoRegionByCode(ctx context.Context, req *geov1.GetGeoRegionByCodeReq) (*geov1.GetGeoRegionByCodeResp, error) {
	regionTrees, err := g.getGeoRegionsByCca2(ctx, g.db, req.CountryCca2)
	if err != nil {
		return nil, err
	}
	if req.Code != nil {
		region := g.findRegion(regionTrees, req.GetCode())
		if region != nil {
			regionTrees = []*geov1.GeoRegion{region}
		}
	}
	if regionTrees != nil {
		g.limitTreeLevel(regionTrees, req.WithChildren, req.ChildrenLevel, req.Lang)
	}
	return &geov1.GetGeoRegionByCodeResp{Regions: regionTrees}, nil
}

func (g *GeoService) CheckGeoRegion(ctx context.Context, req *geov1.CheckGeoRegionReq) (*geov1.CheckGeoRegionResp, error) {
	var ok bool
	regionTree, err := g.getGeoRegionsByCca2(ctx, g.db, req.CountryCca2)
	if err != nil {
		return nil, err
	}
	if province := g.findRegion(regionTree, req.ProvinceCode); province != nil {
		if city := g.findRegion([]*geov1.GeoRegion{province}, req.CityCode); city != nil {
			if district := g.findRegion([]*geov1.GeoRegion{city}, req.DistrictCode); district != nil {
				ok = true
			}
		}
	}
	return &geov1.CheckGeoRegionResp{Ok: ok}, nil
}

func (g *GeoService) getCca2ByRegionID(ctx context.Context, tx *query.Query, regionID int64) (cca2 string, err error) {
	q := tx.GeoRegion
	m, err := q.WithContext(ctx).Select(q.CountryCca2).Where(q.ID.Eq(regionID)).First()
	if err != nil {
		if db.IsDBNotFoundErr(err) {
			return "", ecode.ErrRegionCodeNotExists
		}
		return "", err
	}
	return m.CountryCca2, nil
}

func (g *GeoService) getRegionsByCca2(ctx context.Context, tx *query.Query, cca2 string) (regions []*geov1.GeoRegion, err error) {
	q := tx.GeoRegion
	modes, err := q.WithContext(ctx).Where(q.CountryCca2.Eq(cca2)).Find()
	if err != nil {
		return nil, err
	}
	if len(modes) == 0 {
		return nil, ecode.ErrCountryRegionsNotImported
	}
	regions = make([]*geov1.GeoRegion, 0, len(modes))
	for _, m := range modes {
		regions = append(regions, pack.RegionModelToProto(m))
	}
	return regions, nil
}

func (g *GeoService) getGeoRegionsByCca2(ctx context.Context, tx *query.Query, cca2 string) ([]*geov1.GeoRegion, error) {
	regionTrees, err := g.cache.GetRegionsByCountryCca2(ctx, cca2)
	if err == nil {
		return regionTrees, nil
	}
	if !cache.IsCacheNotFoundErr(err) {
		return nil, err
	}
	key := g.cache.GetCountryCca2RegionKey(cca2)
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	res, err := g.getRegionsByCca2(ctx, tx, cca2)
	if err != nil {
		return nil, err
	}
	regionTrees = g.buildRegionTree(res)
	if err := g.cache.SetRegionsByCountryCca2(ctx, cca2, regionTrees); err != nil {
		return nil, err
	}
	return regionTrees, nil
}

func (g *GeoService) addRegions(ctx context.Context, regions []*geov1.RegionCode) ([]*geov1.AddRegionResult, error) {
	results := make([]*geov1.AddRegionResult, len(regions))
	countryMappings, err := g.regionCodesToCca2Mappings(regions)
	if err != nil {
		return nil, err
	}
	for country, rs := range countryMappings {
		key := g.cache.GetCountryCca2RegionKey(country)
		identifier, err := g.cache.Lock(ctx, key)
		if err != nil {
			return nil, err
		}
		err = g.db.Transaction(func(tx *query.Query) error {
			for i, region := range rs {
				results[i] = &geov1.AddRegionResult{
					CountryCca2: region.GetCountryCca2(),
					NameEnUs:    pack.GetLangName(region.MultiLang, enumsv1.Language_LANGUAGE_EN_US),
					NameZhCn:    pack.GetLangName(region.MultiLang, enumsv1.Language_LANGUAGE_ZH_CN),
					Source:      trans.Deref(region.Source),
					Level:       trans.Deref(region.Level),
					ParentCode:  trans.Deref(region.ParentCode),
					Code:        trans.Deref(region.Code),
					Result:      resultOK,
				}
				m, err := g.createRegionCodeIfNotExists(ctx, tx, region)
				if err != nil {
					if errors.Is(err, ecode.ErrRegionCodeExists) {
						results[i].Result = err.Error()
						continue
					}
					return err
				}
				results[i].Id = m.ID
			}
			return g.cache.DeleteRegionsByCountryCca2(ctx, country)
		})
		g.cache.Unlock(ctx, key, identifier)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (g *GeoService) regionCodesToCca2Mappings(regions []*geov1.RegionCode) (mappings map[string][]*geov1.RegionCode, err error) {
	mappings = make(map[string][]*geov1.RegionCode)
	for _, region := range regions {
		if region.CountryCca2 == nil {
			return nil, ecode.ErrCountryCca2NotExist
		}
		mappings[*region.CountryCca2] = append(mappings[*region.CountryCca2], region)
	}
	return mappings, nil
}

func (g *GeoService) checkRegionIDExists(ctx context.Context, tx *query.Query, id int64) (bool, error) {
	q := tx.GeoRegion
	_, err := q.WithContext(ctx).Select(q.ID).Where(q.ID.Eq(id)).Take()
	if err != nil {
		if db.IsDBNotFoundErr(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g *GeoService) checkRegionCodeExists(ctx context.Context, tx *query.Query, countryCca2, code string) (bool, error) {
	q := tx.GeoRegion
	_, err := tx.WithContext(ctx).GeoRegion.Select(q.ID).Where(q.CountryCca2.Eq(countryCca2), q.Code.Eq(code)).Take()
	if err != nil {
		if db.IsDBNotFoundErr(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g *GeoService) createRegionCodeIfNotExists(
	ctx context.Context,
	tx *query.Query,
	code *geov1.RegionCode,
) (*model.GeoRegion, error) {
	cc := trans.Deref(code.CountryCca2)
	c := trans.Deref(code.Code)

	exists, err := g.checkRegionCodeExists(ctx, tx, cc, c)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ecode.ErrRegionCodeExists
	}

	m := pack.RegionCodeProtoToModels(code)
	if err := tx.GeoRegion.WithContext(ctx).Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

// buildRegionTree 构建地区树形结构
func (g *GeoService) buildRegionTree(regions []*geov1.GeoRegion) []*geov1.GeoRegion {
	if len(regions) == 0 {
		return nil
	}

	nodeMap := make(map[string]*geov1.GeoRegion, len(regions))
	for _, region := range regions {
		region.Children = nil // 清空旧的
		nodeMap[region.Code] = region
	}

	// 第二步：连接父子关系
	var roots []*geov1.GeoRegion
	for _, region := range regions {
		if region.ParentCode == "" || region.ParentCode == region.Code {
			roots = append(roots, region)
			continue
		}

		parent := nodeMap[region.ParentCode]
		if parent != nil {
			parent.Children = append(parent.Children, region)
		} else {
			// 父节点不存在，也当作根节点处理
			roots = append(roots, region)
		}
	}

	return roots
}

func (g *GeoService) findRegion(nodeTrees []*geov1.GeoRegion, code string) *geov1.GeoRegion {
	for _, node := range nodeTrees {
		if node.Code == code {
			return node
		}
		if child := g.findRegion(node.Children, code); child != nil {
			return child
		}
	}
	return nil
}

// limitTreeLevel 限制树的层级
func (g *GeoService) limitTreeLevel(
	nodes []*geov1.GeoRegion,
	withChildren bool,
	maxLevel enumsv1.RegionLevel,
	lang enumsv1.Language,
) {
	// 递归处理子节点
	for _, node := range nodes {
		node.Name = pack.GetLangName(node.MultiLang, lang)
		node.MultiLang = nil
		if !withChildren {
			continue
		}
		if node.Level >= maxLevel {
			node.Children = nil
		} else if len(node.Children) > 0 {
			g.limitTreeLevel(node.Children, withChildren, maxLevel, lang)
		}
	}
}
