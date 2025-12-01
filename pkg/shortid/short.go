package shortid

import (
	"github.com/sqids/sqids-go"
)

// Generator
// 文档：https://sqids.org/go
type Generator struct {
	cli *sqids.Sqids
}

type Config struct {
	Alphabet  string
	MinLength uint8
	Blocklist []string
}

func NewShortIdGenerator(cfg *Config) (generator *Generator, err error) {
	cli, err := sqids.New(sqids.Options{
		Alphabet:  cfg.Alphabet,
		MinLength: cfg.MinLength,
		Blocklist: cfg.Blocklist,
	})
	if err != nil {
		return nil, err
	}
	return &Generator{cli: cli}, nil
}

func (s *Generator) Encode(numbers []uint64) (id string, err error) {
	return s.cli.Encode(numbers)
}

func (s *Generator) Decode(id string) []uint64 {
	return s.cli.Decode(id)
}
