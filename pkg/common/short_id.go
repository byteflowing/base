package common

import "github.com/byteflowing/go-common/idx"

type ShortIDGenerator struct {
	shortIDGen  *idx.ShortIDGenerator
	globalIDGen *idx.GlobalIDGenerator
}

func NewShortIDGenerator(globalID *idx.GlobalIDGenerator, opts *idx.ShotIDGeneratorOpts) *ShortIDGenerator {
	shortIDGenerator, err := idx.NewShortIdGenerator(opts)
	if err != nil {
		panic(err)
	}
	return &ShortIDGenerator{
		shortIDGen:  shortIDGenerator,
		globalIDGen: globalID,
	}
}

func (s *ShortIDGenerator) GetID() string {
	// 这里基本上不会出错
	globalID, err := s.globalIDGen.NextID()
	if err != nil {
		panic(err)
	}
	// 这里不太会报错
	id, err := s.shortIDGen.Encode([]uint64{uint64(globalID)})
	if err != nil {
		panic(err)
	}
	return id
}

func (s *ShortIDGenerator) Decode(id string) []uint64 {
	return s.shortIDGen.Decode(id)
}
