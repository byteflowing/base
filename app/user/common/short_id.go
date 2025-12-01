package common

import (
	"context"
	"sync"

	globalId "github.com/byteflowing/base/app/global_id/service"
	"github.com/byteflowing/base/pkg/shortid"
	globalidv1 "github.com/byteflowing/proto/gen/go/global_id/v1"
)

var (
	shortIDOnce sync.Once
	shortID     *ShortID
)

type ShortID struct {
	globalIDService *globalId.GlobalIDService
	shortId         *shortid.Generator
}

func NewShortID(globalIDService *globalId.GlobalIDService, shortId *shortid.Generator) *ShortID {
	shortIDOnce.Do(func() {
		shortID = &ShortID{
			globalIDService: globalIDService,
			shortId:         shortId,
		}
	})
	return shortID
}

func (s *ShortID) GetShortID(ctx context.Context) (string, error) {
	res, err := s.globalIDService.GetId(ctx, &globalidv1.GetIdReq{})
	if err != nil {
		return "", err
	}
	return s.shortId.Encode([]uint64{uint64(res.Id)})
}
