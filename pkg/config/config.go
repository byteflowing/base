package config

import (
	"bytes"
	"os"
	"regexp"
	"strings"

	"github.com/byteflowing/base/pkg/jsonx"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ReadConfig 读取配置文件
// 支持环境变量 ${VAR:-default}
func ReadConfig(file string, config interface{}) (err error) {
	v := viper.New()
	if err := readConfigAndExpendEvn(v, file); err != nil {
		return err
	}
	return v.Unmarshal(config)
}

// ReadProtoConfig 读取配置文件，并将配置文件写入proto生成的结构中
// 支持环境变量 ${VAR:-default}
func ReadProtoConfig(file string, msg proto.Message) (err error) {
	v := viper.New()
	if err := readConfigAndExpendEvn(v, file); err != nil {
		return err
	}
	allSettings := v.AllSettings()
	data, err := jsonx.Marshal(allSettings)
	if err != nil {
		return err
	}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	return unmarshaler.Unmarshal(data, msg)
}

func expandEnvWithDefault(file string) ([]byte, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`\$\{([^}:]+)(?:(:-)([^}]*))?\}`)
	expanded := re.ReplaceAllStringFunc(string(raw), func(m string) string {
		matches := re.FindStringSubmatch(m)
		if len(matches) < 2 {
			return m // 安全兜底
		}
		key := matches[1] // 变量名
		def := ""
		if len(matches) >= 4 {
			def = matches[3] // 默认值，可能为空
		}
		val, ok := os.LookupEnv(key)
		if !ok {
			return def // 未定义 → 用默认值或空
		}
		if matches[2] == ":-" && val == "" {
			return def // :-default 且空 → 用默认值
		}
		return val
	})
	return []byte(expanded), nil
}

func readConfigAndExpendEvn(v *viper.Viper, file string) error {
	v.SetConfigFile(file)
	v.AutomaticEnv()
	v.AllowEmptyEnv(true)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	expended, err := expandEnvWithDefault(v.ConfigFileUsed())
	if err != nil {
		return err
	}
	return v.ReadConfig(bytes.NewReader(expended))
}
