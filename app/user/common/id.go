package common

import (
	"context"
	"sync"

	"github.com/bytedance/gopkg/lang/fastrand"
	globalId "github.com/byteflowing/base/app/global_id/service"
	"github.com/byteflowing/base/pkg/shortid"
	globalidv1 "github.com/byteflowing/proto/gen/go/global_id/v1"
)

var (
	idOnce    sync.Once
	idService *IDService
)

type IDService struct {
	globalIDService *globalId.GlobalIDService
	shortId         *shortid.Generator
}

func NewIDService(globalIDService *globalId.GlobalIDService, shortId *shortid.Generator) *IDService {
	idOnce.Do(func() {
		idService = &IDService{
			globalIDService: globalIDService,
			shortId:         shortId,
		}
	})
	return idService
}

func (s *IDService) GetShortID(ctx context.Context) (string, error) {
	res, err := s.globalIDService.GetId(ctx, &globalidv1.GetIdReq{})
	if err != nil {
		return "", err
	}
	return s.shortId.Encode([]uint64{uint64(res.Id), fastrand.Uint64()})
}

func (s *IDService) GetGlobalID(ctx context.Context) (int64, error) {
	res, err := s.globalIDService.GetId(ctx, &globalidv1.GetIdReq{})
	if err != nil {
		return 0, err
	}
	return res.Id, nil
}
