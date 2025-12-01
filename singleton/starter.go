package singleton

import (
	"sync"

	"github.com/byteflowing/base/pkg/srvmgr"
)

var (
	starterOnce sync.Once
	starterMgr  *srvmgr.ServiceManager
)

// GetStarterMgr 管理有start和stop方法的单利结构，在系统启动时统一start，在系统关闭时统一stop
func GetStarterMgr() *srvmgr.ServiceManager {
	if starterMgr == nil {
		return nil
	}
	return getStarterMgr()
}

func addStarter(srv srvmgr.Service) {
	getStarterMgr().Add(srv)
}

func getStarterMgr() *srvmgr.ServiceManager {
	starterOnce.Do(func() {
		starterMgr = srvmgr.NewServiceManager()
	})
	return starterMgr
}
