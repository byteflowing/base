package common

import (
	"time"

	configv1 "github.com/byteflowing/base/gen/config/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	"github.com/byteflowing/go-common/idx"
)

const (
	machineID = 1
)

type GlobalIdGenerator interface {
	GetID() (int64, error)
}

func NewGlobalIdGenerator(c *configv1.GlobalId) GlobalIdGenerator {
	if c.Mode == enumsv1.GlobalIdMode_GLOBAL_ID_MODE_LOCAL {
		return &LocalGlobalIdGenerator{
			idGen: newLocalGlobalId(c.StartTime.AsTime()),
		}
	}
	panic("invalid globalId mode")
}

type LocalGlobalIdGenerator struct {
	idGen *idx.GlobalIDGenerator
}

func (l *LocalGlobalIdGenerator) GetID() (int64, error) {
	return l.idGen.NextID()
}

// 这里只能单机使用
func newLocalGlobalId(startTime time.Time) *idx.GlobalIDGenerator {
	generator, err := idx.NewGlobalIDGenerator(&idx.GlobalIDOpts{
		StartTime:      startTime,
		MachineID:      getMachineID,
		CheckMachineID: checkMachineID,
	})
	if err != nil {
		panic(err)
	}
	return generator
}

func getMachineID() (int, error) {
	return machineID, nil
}

func checkMachineID(id int) bool {
	return id == machineID
}
