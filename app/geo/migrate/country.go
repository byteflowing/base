package migrate

import (
	"context"

	"go.uber.org/zap"

	"github.com/byteflowing/base/app/geo/dal/model"
	"github.com/byteflowing/base/pkg/config"
	"github.com/byteflowing/base/pkg/jsonx"
	"github.com/byteflowing/base/pkg/logx"
	"github.com/byteflowing/base/pkg/utils/trans"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
)

func (m *Migrate) MigrateCountries(filePath string) error {
	ctx := context.Background()
	q := m._query.GeoCountry
	if _, err := q.WithContext(ctx).Where(q.ID.Gt(0)).Take(); err == nil {
		logx.Warn("country codes already exist, so ignore import")
		return nil
	}
	logx.Warn("-------------------------importing country codes started------------------------------------------")
	defer logx.Warn("-------------------------importing country codes ended--------------------------------------")
	logx.Warn("importing country codes file", zap.String("file", filePath))
	var countries []*geov1.ExternalCountry
	if err := config.ReadConfig(filePath, &countries); err != nil {
		return err
	}
	if len(countries) == 0 {
		logx.Warn("no country codes found, so ignore import")
		return nil
	}
	var countryCodes []*model.GeoCountry
	for _, country := range countries {
		lang := toMultiLang(country.Translations)
		lang[enumsv1.Language_LANGUAGE_EN_US.String()] = country.Name.Common
		multiLang, err := jsonx.MarshalToString(lang)
		if err != nil {
			return err
		}
		countryCodes = append(countryCodes, &model.GeoCountry{
			Cca2:         country.Cca2,
			Cca3:         country.Cca3,
			Ccn3:         country.Ccn3,
			Flag:         country.Flag,
			Continent:    country.Region,
			SubContinent: country.Subregion,
			MultiLang:    trans.Ref(multiLang),
			Independent:  country.Independent,
			IsActive:     trans.Ref(true),
		})
	}
	if len(countryCodes) == 0 {
		logx.Warn("no country codes found, so ignore import")
	}
	return q.WithContext(ctx).Create(countryCodes...)
}

func toMultiLang(trans *geov1.ExternalCountry_Translations) map[string]string {
	lang := make(map[string]string)
	if trans.Zho != nil {
		lang[enumsv1.Language_LANGUAGE_ZH_CN.String()] = trans.Zho.Common
	}
	if trans.Jpn != nil {
		lang[enumsv1.Language_LANGUAGE_JA_JP.String()] = trans.Jpn.Common
	}
	if trans.Kor != nil {
		lang[enumsv1.Language_LANGUAGE_KO_KR.String()] = trans.Kor.Common
	}
	if trans.Fra != nil {
		lang[enumsv1.Language_LANGUAGE_FR_FR.String()] = trans.Fra.Common
	}
	if trans.Deu != nil {
		lang[enumsv1.Language_LANGUAGE_DE_DE.String()] = trans.Deu.Common
	}
	if trans.Spa != nil {
		lang[enumsv1.Language_LANGUAGE_ES_ES.String()] = trans.Spa.Common
	}
	if trans.Por != nil {
		lang[enumsv1.Language_LANGUAGE_PT_PT.String()] = trans.Por.Common
	}
	if trans.Rus != nil {
		lang[enumsv1.Language_LANGUAGE_RU_RU.String()] = trans.Rus.Common
	}
	if trans.Ara != nil {
		lang[enumsv1.Language_LANGUAGE_AR_SA.String()] = trans.Ara.Common
	}
	if trans.Tur != nil {
		lang[enumsv1.Language_LANGUAGE_TR_TR.String()] = trans.Tur.Common
	}
	if trans.Ita != nil {
		lang[enumsv1.Language_LANGUAGE_IT_IT.String()] = trans.Ita.Common
	}
	if trans.Nld != nil {
		lang[enumsv1.Language_LANGUAGE_NL_NL.String()] = trans.Nld.Common
	}
	if trans.Pol != nil {
		lang[enumsv1.Language_LANGUAGE_PL_PL.String()] = trans.Pol.Common
	}
	if trans.Swe != nil {
		lang[enumsv1.Language_LANGUAGE_SV_SE.String()] = trans.Swe.Common
	}
	if trans.Per != nil {
		lang[enumsv1.Language_LANGUAGE_FA_IR.String()] = trans.Per.Common
	}
	return lang
}
