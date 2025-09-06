package common

import (
	"github.com/byteflowing/go-common/idx"
	idxv1 "github.com/byteflowing/proto/gen/go/idx/v1"
)

type ShortIDGenerator struct {
	shortIDGen  *idx.ShortIDGenerator
	globalIDGen GlobalIdGenerator
}

func NewShortIDGenerator(globalIDGen GlobalIdGenerator, c *idxv1.ShortIdConfig) *ShortIDGenerator {
	shortIDGenerator, err := idx.NewShortIdGenerator(c)
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
