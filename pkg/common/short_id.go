package common

import (
	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/idx"
)

type ShortIDGenerator struct {
	shortIDGen  *idx.ShortIDGenerator
	globalIDGen GlobalIdGenerator
}

func NewShortIDGenerator(globalIDGen GlobalIdGenerator, c *configv1.ShortId) *ShortIDGenerator {
	shortIDGenerator, err := idx.NewShortIdGenerator(&idx.ShotIDGeneratorOpts{
		Alphabet:  c.Alphabet,
		MinLength: uint8(c.MinLength),
		Blocklist: c.BlockList,
	})
	if err != nil {
		panic(err)
	}
	return &ShortIDGenerator{
		shortIDGen:  shortIDGenerator,
		globalIDGen: globalIDGen,
	}
}

func (s *ShortIDGenerator) GetID() (id string, err error) {
	globalID, err := s.globalIDGen.GetID()
	if err != nil {
		return "", err
	}
	return s.shortIDGen.Encode([]uint64{uint64(globalID)})
}

func (s *ShortIDGenerator) Decode(id string) []uint64 {
	return s.shortIDGen.Decode(id)
}
