package pack

import (
	"github.com/byteflowing/base/pkg/jsonx"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
)

func GetLangName(names map[string]string, lang enumsv1.Language) string {
	if name, ok := names[lang.String()]; ok {
		return name
	}
	if enName, ok := names[enumsv1.Language_LANGUAGE_EN_US.String()]; ok {
		return enName
	}
	return ""
}

func MultiLangToJsonString(lang map[string]string) *string {
	if len(lang) == 0 {
		return nil
	}
	b, err := jsonx.MarshalToString(lang)
	if err != nil {
		panic(err)
	}
	return &b
}

func MultiLangFromJsonString(jsonStr *string) map[string]string {
	if jsonStr == nil {
		return nil
	}
	multiLang := make(map[string]string)
	if err := jsonx.UnmarshalFromString(*jsonStr, &multiLang); err != nil {
		panic(err)
	}
	return multiLang
}
