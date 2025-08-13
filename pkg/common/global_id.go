package common

import (
	"time"

	"github.com/byteflowing/go-common/idx"
)

const (
	machineID = 1
)

// NewGlobalID 生成全球唯一id
// 这里只能单机使用
func NewGlobalID() *idx.GlobalIDGenerator {
	t := time.Date(2025, 8, 8, 8, 8, 8, 8, time.Local)
	generator, err := idx.NewGlobalIDGenerator(&idx.GlobalIDOpts{
		StartTime:      t,
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
