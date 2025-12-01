package service

import (
	"context"

	"github.com/byteflowing/base/pkg/globalid"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	globalidv1 "github.com/byteflowing/proto/gen/go/global_id/v1"
)

type GlobalIDService struct {
	idGen *globalid.Generator

	globalidv1.UnimplementedGlobalIdServiceServer
}

// NewGlobalIDService 创建globaId generator实例
// 通过GetId接口获取单个全局id
// 通过GetIds接口获取多个全局id
// generator是sonyflake的封装，Decompose ToTime等方法，可以使用pkg/globalid中的对应的方法
func NewGlobalIDService(cfg *configv1.Config) *GlobalIDService {
	c := cfg.GlobalId
	generator, err := globalid.NewGlobalIDGenerator(&globalid.Config{
		StartTime:      c.StartTime.AsTime(),
		MachineID:      getMachineID(c),
		CheckMachineID: checkMachineID(c),
	})
	if err != nil {
		panic(err)
	}
	return &GlobalIDService{idGen: generator}
}

func (g *GlobalIDService) GetId(ctx context.Context, req *globalidv1.GetIdReq) (*globalidv1.GetIdResp, error) {
	_id, err := g.idGen.NextID()
	if err != nil {
		return nil, err
	}
	return &globalidv1.GetIdResp{Id: _id}, nil
}

func (g *GlobalIDService) GetIds(ctx context.Context, req *globalidv1.GetIdsReq) (*globalidv1.GetIdsResp, error) {
	var ids []int64
	for i := 0; i < int(req.Num); i++ {
		_id, err := g.idGen.NextID()
		if err != nil {
			return nil, err
		}
		ids = append(ids, _id)
	}
	return &globalidv1.GetIdsResp{Ids: ids}, nil
}

func getMachineID(cfg *globalidv1.GlobalIdConfig) func() (int, error) {
	return func() (int, error) {
		return int(cfg.MachineId), nil
	}
}

func checkMachineID(cfg *globalidv1.GlobalIdConfig) func(id int) bool {
	return func(id int) bool {
		return id == int(cfg.MachineId)
	}
}
