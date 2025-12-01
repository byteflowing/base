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
	"github.com/byteflowing/go-common/trans"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
)

func (g *GeoService) AddCountryCode(ctx context.Context, req *geov1.AddCountryCodeReq) (*geov1.AddCountryCodeResp, error) {
	results, err := g.addCountryCodes(ctx, []*geov1.CountryCode{req.GetCountryCode()})
	if err != nil {
		return nil, err
	}
	var result *geov1.AddCountryCodeResult
	if len(results) > 0 {
		result = results[0]
	}
	return &geov1.AddCountryCodeResp{Result: result}, nil
}

func (g *GeoService) AddCountryCodes(ctx context.Context, req *geov1.AddCountryCodesReq) (*geov1.AddCountryCodesResp, error) {
	results, err := g.addCountryCodes(ctx, req.GetCountryCodes())
	if err != nil {
		return nil, err
	}
	return &geov1.AddCountryCodesResp{Results: results}, nil
}

func (g *GeoService) UpdateCountryCode(ctx context.Context, req *geov1.UpdateCountryCodeReq) (*geov1.UpdateCountryCodeResp, error) {
	key := g.cache.GetCountryKey()
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	err = g.db.Transaction(func(tx *query.Query) error {
		exists, err := g.checkCountryIDExists(ctx, g.db, req.Id)
		if err != nil {
			return err
		}
		if !exists {
			return ecode.ErrCountryCodeNotExist
		}
		m := pack.CountryCodeProtoToModels(req.Info)
		q := tx.GeoCountry
		if _, err = q.WithContext(ctx).Where(q.ID.Eq(req.Id)).Updates(m); err != nil {
			return err
		}
		return g.cache.DeleteAllCountries(ctx)
	})
	if err != nil {
		return nil, err
	}
	return &geov1.UpdateCountryCodeResp{}, err
}

func (g *GeoService) DeleteCountryCode(ctx context.Context, req *geov1.DeleteCountryCodeReq) (*geov1.DeleteCountryCodeResp, error) {
	key := g.cache.GetCountryKey()
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	err = g.db.Transaction(func(tx *query.Query) error {
		exists, err := g.checkCountryIDExists(ctx, tx, req.Id)
		if err != nil {
			return err
		}
		if !exists {
			return ecode.ErrCountryCodeNotExist
		}
		q := tx.GeoCountry
		if _, err := q.WithContext(ctx).Where(q.ID.Eq(req.Id)).Delete(); err != nil {
			return err
		}
		return g.cache.DeleteAllCountries(ctx)
	})
	if err != nil {
		return nil, err
	}
	return &geov1.DeleteCountryCodeResp{}, nil
}

func (g *GeoService) GetAllCountries(ctx context.Context, req *geov1.GetAllCountriesReq) (*geov1.GetAllCountriesResp, error) {
	resp, err := g.cache.GetAllCountries(ctx, req)
	if err == nil {
		return resp, nil
	}
	if !cache.IsCacheNotFoundErr(err) {
		return nil, err
	}
	key := g.cache.GetCountryKey()
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	exists, err := g.cache.CheckExists(ctx, key)
	if err != nil {
		return nil, err
	}
	if exists {
		return g.cache.GetAllCountries(ctx, req)
	}
	countries, err := g.getAllCountries(ctx, g.db)
	if err != nil {
		return nil, err
	}
	if err := g.cache.SetAllCountries(ctx, countries); err != nil {
		return nil, err
	}
	return g.cache.GetAllCountries(ctx, req)
}

func (g *GeoService) GetCountryById(ctx context.Context, req *geov1.GetCountryByIdReq) (*geov1.GetCountryByIdResp, error) {
	resp, err := g.cache.GetCountryByID(ctx, req)
	if err == nil {
		return resp, nil
	}
	if !cache.IsCacheNotFoundErr(err) {
		return nil, err
	}
	key := g.cache.GetCountryKey()
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	exists, err := g.cache.CheckExists(ctx, key)
	if err != nil {
		return nil, err
	}
	if exists {
		return g.cache.GetCountryByID(ctx, req)
	}
	countries, err := g.getAllCountries(ctx, g.db)
	if err != nil {
		return nil, err
	}
	if err := g.cache.SetAllCountries(ctx, countries); err != nil {
		return nil, err
	}
	return g.cache.GetCountryByID(ctx, req)
}

func (g *GeoService) GetCountryByCca2(ctx context.Context, req *geov1.GetCountryByCca2Req) (*geov1.GetCountryByCca2Resp, error) {
	resp, err := g.cache.GetCountryByCca2(ctx, req)
	if err == nil {
		return resp, nil
	}
	if !cache.IsCacheNotFoundErr(err) {
		return nil, err
	}
	key := g.cache.GetCountryKey()
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	exists, err := g.cache.CheckExists(ctx, key)
	if err != nil {
		return nil, err
	}
	if exists {
		return g.cache.GetCountryByCca2(ctx, req)
	}
	countries, err := g.getAllCountries(ctx, g.db)
	if err != nil {
		return nil, err
	}
	if err := g.cache.SetAllCountries(ctx, countries); err != nil {
		return nil, err
	}
	return g.cache.GetCountryByCca2(ctx, req)
}

func (g *GeoService) getAllCountries(ctx context.Context, tx *query.Query) (codes []*geov1.GeoCountry, err error) {
	q := tx.GeoCountry
	modes, err := q.WithContext(ctx).Where(q.ID.Gt(0)).Find()
	if err != nil {
		return nil, err
	}
	if len(modes) == 0 {
		return nil, ecode.ErrCountryCodeNotImported
	}
	codes = make([]*geov1.GeoCountry, 0, len(modes))
	for _, m := range modes {
		codes = append(codes, pack.CountryModelToGeoProto(m))
	}
	return codes, nil
}

func (g *GeoService) countryCodesToCca2Mappings(codes []*geov1.GeoCountry) map[string]*geov1.GeoCountry {
	mappings := make(map[string]*geov1.GeoCountry, len(codes))
	for _, code := range codes {
		mappings[code.Cca2] = code
	}
	return mappings
}

func (g *GeoService) addCountryCodes(ctx context.Context, countryCodes []*geov1.CountryCode) ([]*geov1.AddCountryCodeResult, error) {
	key := g.cache.GetCountryKey()
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	results := make([]*geov1.AddCountryCodeResult, len(countryCodes))
	err = g.db.Transaction(func(tx *query.Query) error {
		for i, c := range countryCodes {
			results[i] = &geov1.AddCountryCodeResult{
				Cca2:     c.GetCca2(),
				NameEnUs: pack.GetLangName(c.MultiLang, enumsv1.Language_LANGUAGE_EN_US),
				NameZhCn: pack.GetLangName(c.MultiLang, enumsv1.Language_LANGUAGE_ZH_CN),
				Result:   resultOK,
			}
			m, err := g.createCountryCodeIfNotExists(ctx, tx, c)
			if err != nil {
				if errors.Is(err, ecode.ErrCountryCodeExists) {
					results[i].Result = err.Error()
					continue
				}
				return err
			}
			results[i].Id = m.ID
		}
		return g.cache.DeleteAllCountries(ctx)
	})
	return results, err
}

func (g *GeoService) checkCountryCodeExists(ctx context.Context, tx *query.Query, cca2 string) (bool, error) {
	q := tx.GeoCountry
	_, err := q.WithContext(ctx).Select(q.ID).Where(q.Cca2.Eq(cca2)).Take()
	if err != nil {
		if db.IsDBNotFoundErr(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g *GeoService) checkCountryIDExists(ctx context.Context, tx *query.Query, id int64) (bool, error) {
	q := tx.GeoCountry
	_, err := q.WithContext(ctx).Select(q.ID).Where(q.ID.Eq(id)).Take()
	if err != nil {
		if db.IsDBNotFoundErr(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g *GeoService) createCountryCodeIfNotExists(
	ctx context.Context,
	tx *query.Query,
	code *geov1.CountryCode,
) (*model.GeoCountry, error) {
	cc := trans.Deref(code.Cca2)

	exists, err := g.checkCountryCodeExists(ctx, tx, cc)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ecode.ErrCountryCodeExists
	}

	m := pack.CountryCodeProtoToModels(code)
	if err := tx.GeoCountry.WithContext(ctx).Create(m); err != nil {
		return nil, err
	}
	return m, nil
}
