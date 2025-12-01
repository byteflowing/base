package globalid

import (
	"errors"
	"time"

	"github.com/sony/sonyflake/v2"
)

// Generator 雪花算法的sony实现
// 文档：https://github.com/sony/sonyflake
type Generator struct {
	generator *sonyflake.Sonyflake
}

type Config struct {
	// 开始时间建议设置为项目开始时间
	// 不要设置未来时间，会导致时间回拨等问题
	// 时间一旦确定请勿随意更改，建议硬编码到程序代码中
	StartTime time.Time
	// 可以使用自增id等等方法
	MachineID func() (int, error)
	// [可选] 用于检测machine id是否重复
	CheckMachineID func(int) bool
}

func NewGlobalIDGenerator(cfg *Config) (sf *Generator, err error) {
	st := sonyflake.Settings{
		StartTime:      cfg.StartTime,
		MachineID:      cfg.MachineID,
		CheckMachineID: cfg.CheckMachineID,
	}
	if cfg.MachineID == nil {
		return nil, errors.New("no machine id")
	}
	if cfg.CheckMachineID == nil {
		return nil, errors.New("no check machine id")
	}
	flake, err := sonyflake.New(st)
	if err != nil {
		return nil, err
	}
	return &Generator{generator: flake}, nil
}

// NextID 生成全局ID
// 每 10 毫秒最多可以生成 256 个 ID
func (s *Generator) NextID() (id int64, err error) {
	return s.generator.NextID()
}

// Decompose 解析出id的组成部分
// id：原始的完整 ID。
// msb：最高有效位，一般为 0。
// time：时间戳部分，表示从 StartTime 开始到 ID 生成的时间。
// sequence：序列号，用于防止同一时间戳内的 ID 冲突。
// machine-id：机器 ID，标识生成该 ID 的机器。
func (s *Generator) Decompose(id int64) map[string]int64 {
	return s.generator.Decompose(id)
}

// ToTime 返回生成该id的时间
func (s *Generator) ToTime(id int64) time.Time {
	return s.generator.ToTime(id)
}
