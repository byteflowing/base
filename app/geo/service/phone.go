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

func (g *GeoService) AddPhoneCode(ctx context.Context, req *geov1.AddPhoneCodeReq) (*geov1.AddPhoneCodeResp, error) {
	results, err := g.addPhoneCodes(ctx, []*geov1.PhoneCode{req.GetCode()})
	if err != nil {
		return nil, err
	}
	var result *geov1.AddPhoneCodeResult
	if len(results) > 0 {
		result = results[0]
	}
	return &geov1.AddPhoneCodeResp{Result: result}, nil
}

func (g *GeoService) AddPhoneCodes(ctx context.Context, req *geov1.AddPhoneCodesReq) (*geov1.AddPhoneCodesResp, error) {
	results, err := g.addPhoneCodes(ctx, req.GetPhoneCodes())
	if err != nil {
		return nil, err
	}
	return &geov1.AddPhoneCodesResp{Results: results}, nil
}

func (g *GeoService) UpdatePhoneCode(ctx context.Context, req *geov1.UpdatePhoneCodeReq) (*geov1.UpdatePhoneCodeResp, error) {
	key := g.cache.GetPhoneKey()
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	err = g.db.Transaction(func(tx *query.Query) error {
		exists, err := g.checkPhoneIDExists(ctx, tx, req.Id)
		if err != nil {
			return err
		}
		if !exists {
			return ecode.ErrPhoneCodeNotExist
		}
		m := pack.PhoneCodeProtoToModel(req.Code)
		q := tx.GeoPhoneCode
		if _, err = q.WithContext(ctx).Where(q.ID.Eq(req.Id)).Updates(m); err != nil {
			return err
		}
		return g.cache.DeleteAllPhoneCodes(ctx)
	})
	if err != nil {
		return nil, err
	}
	return &geov1.UpdatePhoneCodeResp{}, nil
}

func (g *GeoService) DeletePhoneCode(ctx context.Context, req *geov1.DeletePhoneCodeReq) (*geov1.DeletePhoneCodeResp, error) {
	key := g.cache.GetPhoneKey()
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	err = g.db.Transaction(func(tx *query.Query) error {
		q := tx.GeoPhoneCode
		if _, err := q.WithContext(ctx).Where(q.ID.Eq(req.Id)).Delete(); err != nil {
			return err
		}
		return g.cache.DeleteAllPhoneCodes(ctx)
	})
	if err != nil {
		return nil, err
	}
	return &geov1.DeletePhoneCodeResp{}, nil
}

func (g *GeoService) GetPhoneCodeById(ctx context.Context, req *geov1.GetPhoneCodeByIdReq) (*geov1.GetPhoneCodeByIdResp, error) {
	resp, err := g.cache.GetPhoneCodeByID(ctx, req)
	if err == nil {
		return resp, err
	}
	if !cache.IsCacheNotFoundErr(err) {
		return nil, err
	}
	key := g.cache.GetPhoneKey()
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
		return g.cache.GetPhoneCodeByID(ctx, req)
	}
	phoneCodes, err := g.getAllPhoneCodes(ctx, g.db)
	if err != nil {
		return nil, err
	}
	if err := g.cache.SetAllPhoneCodes(ctx, phoneCodes); err != nil {
		return nil, err
	}
	return g.cache.GetPhoneCodeByID(ctx, req)
}

func (g *GeoService) GetAllPhoneCodes(ctx context.Context, req *geov1.GetAllPhoneCodesReq) (*geov1.GetAllPhoneCodesResp, error) {
	resp, err := g.cache.GetAllPhoneCodes(ctx, req)
	if err == nil {
		return resp, nil
	}
	if !cache.IsCacheNotFoundErr(err) {
		return nil, err
	}
	key := g.cache.GetPhoneKey()
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
		return g.cache.GetAllPhoneCodes(ctx, req)
	}
	phoneCodes, err := g.getAllPhoneCodes(ctx, g.db)
	if err != nil {
		return nil, err
	}
	if err := g.cache.SetAllPhoneCodes(ctx, phoneCodes); err != nil {
		return nil, err
	}
	return g.cache.GetAllPhoneCodes(ctx, req)
}

func (g *GeoService) getAllPhoneCodes(ctx context.Context, tx *query.Query) (codes []*geov1.GeoPhoneCode, err error) {
	q := tx.GeoPhoneCode
	modes, err := q.WithContext(ctx).Where(q.ID.Gt(0)).Find()
	if err != nil {
		return nil, err
	}
	if len(modes) == 0 {
		return nil, ecode.ErrPhoneCodeNotImported
	}
	codes = make([]*geov1.GeoPhoneCode, 0, len(modes))
	for _, m := range modes {
		codes = append(codes, pack.PhoneCodeModelToProto(m))
	}
	return codes, nil
}

func (g *GeoService) addPhoneCodes(ctx context.Context, phoneCodes []*geov1.PhoneCode) ([]*geov1.AddPhoneCodeResult, error) {
	key := g.cache.GetPhoneKey()
	identifier, err := g.cache.Lock(ctx, key)
	if err != nil {
		return nil, err
	}
	defer g.cache.Unlock(ctx, key, identifier)
	results := make([]*geov1.AddPhoneCodeResult, len(phoneCodes))
	err = g.db.Transaction(func(tx *query.Query) error {
		for i, c := range phoneCodes {
			results[i] = &geov1.AddPhoneCodeResult{
				PhoneCode: c.GetPhoneCode(),
				NameEnUs:  pack.GetLangName(c.MultiLang, enumsv1.Language_LANGUAGE_EN_US),
				NameZhCn:  pack.GetLangName(c.MultiLang, enumsv1.Language_LANGUAGE_ZH_CN),
				Result:    resultOK,
			}
			m, err := g.createPhoneCodeIfNotExists(ctx, tx, c)
			if err != nil {
				if errors.Is(err, ecode.ErrPhoneCodeExists) {
					results[i].Result = err.Error()
					continue
				}
				return err
			}
			results[i].Id = m.ID
		}
		return g.cache.DeleteAllPhoneCodes(ctx)
	})
	return results, err
}

func (g *GeoService) checkPhoneCodeExists(ctx context.Context, tx *query.Query, phoneCode, name string) (bool, error) {
	q := tx.GeoPhoneCode
	_, err := tx.WithContext(ctx).GeoPhoneCode.Select(q.ID).Where(q.PhoneCode.Eq(phoneCode), q.Name.Eq(name)).Take()
	if err != nil {
		if db.IsDBNotFoundErr(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g *GeoService) checkPhoneIDExists(ctx context.Context, tx *query.Query, id int64) (bool, error) {
	q := tx.GeoPhoneCode
	_, err := q.WithContext(ctx).Select(q.ID).Where(q.ID.Eq(id)).Take()
	if err != nil {
		if db.IsDBNotFoundErr(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g *GeoService) createPhoneCodeIfNotExists(
	ctx context.Context,
	tx *query.Query,
	code *geov1.PhoneCode,
) (*model.GeoPhoneCode, error) {
	pc := trans.Deref(code.PhoneCode)
	name := trans.Deref(code.Name)
	exists, err := g.checkPhoneCodeExists(ctx, tx, pc, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ecode.ErrPhoneCodeExists
	}
	m := pack.PhoneCodeProtoToModel(code)
	if err := tx.GeoPhoneCode.WithContext(ctx).Create(m); err != nil {
		return nil, err
	}
	return m, nil
}
