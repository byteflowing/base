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

func (m *Migrate) MigratePhoneCode(filePath string) error {
	ctx := context.Background()
	q := m._query.GeoPhoneCode
	if _, err := q.WithContext(ctx).Where(q.ID.Gt(0)).Take(); err == nil {
		logx.Warn("phone codes already exist, so ignore import")
		return nil
	}
	logx.Warn("-------------------------importing phone codes started--------------------------------------------")
	defer logx.Warn("-------------------------importing phone codes ended----------------------------------------")
	logx.Warn("importing phone codes file", zap.String("file", filePath))
	var phones []*geov1.ExternalPhoneCode
	if err := config.ReadConfig(filePath, &phones); err != nil {
		return err
	}
	if len(phones) == 0 {
		logx.Warn("no phone codes found, so ignore import")
		return nil
	}
	var codes []*model.GeoPhoneCode
	for _, phone := range phones {
		lang := make(map[string]string)
		lang[enumsv1.Language_LANGUAGE_EN_US.String()] = phone.EnglishName
		lang[enumsv1.Language_LANGUAGE_ZH_CN.String()] = phone.ChineseName
		multiLang, err := jsonx.MarshalToString(lang)
		if err != nil {
			return err
		}
		codes = append(codes, &model.GeoPhoneCode{
			Name:      phone.EnglishName,
			PhoneCode: phone.PhoneCode,
			MultiLang: trans.Ref(multiLang),
			IsActive:  trans.Ref(true),
		})
	}
	if len(codes) == 0 {
		logx.Warn("no phone codes found, so ignore import")
	}
	return q.WithContext(ctx).Create(codes...)
}
