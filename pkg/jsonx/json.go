package jsonx

import (
	"github.com/bytedance/sonic"
)

func Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

func MarshalToString(v interface{}) (string, error) {
	return sonic.MarshalString(v)
}

func Unmarshal(data []byte, v interface{}) (err error) {
	return sonic.Unmarshal(data, v)
}

func UnmarshalFromString(str string, v interface{}) (err error) {
	return sonic.UnmarshalString(str, v)
}
