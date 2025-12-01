package srvmgr

import (
	"reflect"
	"sync"

	"github.com/byteflowing/base/pkg/thread"
)

const (
	defaultCap = 10
)

type Service interface {
	Start() error
	Stop()
}

// ServiceManager 服务启动管理
// 实现Service的服务，通过Add方法添加，自带去重
// Start和Stop只可调用一次，内部使用once保证
type ServiceManager struct {
	mux       sync.Mutex
	startOnce sync.Once
	stopOnce  sync.Once
	services  []Service
	seen      map[uintptr]struct{}
}

type StartFunc func() error

func (s StartFunc) Start() error {
	return s()
}

func (s StartFunc) Stop() {}

type StopFunc func()

func (s StopFunc) Start() error {
	return nil
}

func (s StopFunc) Stop() {
	s()
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make([]Service, 0, defaultCap),
		seen:     make(map[uintptr]struct{}, defaultCap),
	}
}

func (sm *ServiceManager) Add(srv Service) {
	sm.add(srv)
}

func (sm *ServiceManager) AddStartFunc(sr StartFunc) {
	sm.add(sr)
}

func (sm *ServiceManager) AddStopFunc(sr StopFunc) {
	sm.add(sr)
}

// Start 启动所有通过Add添加的Service
func (sm *ServiceManager) Start() {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	sm.startOnce.Do(func() {
		for _, srv := range sm.services {
			s := srv
			thread.GoUnsafe(func() {
				if err := s.Start(); err != nil {
					panic(err)
				}
			})
		}
	})
}

func (sm *ServiceManager) Stop() {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	sm.stopOnce.Do(func() {
		sg := thread.NewRoutineGroup()
		for _, srv := range sm.services {
			s := srv
			sg.RunSafe(func() {
				s.Stop()
			})
		}
		sg.Wait()
	})
}

func (sm *ServiceManager) add(srv Service) {
	ptr := reflect.ValueOf(srv).Pointer()
	sm.mux.Lock()
	defer sm.mux.Unlock()
	// 重复实例无需添加
	if _, ok := sm.seen[ptr]; ok {
		return
	}
	sm.services = append(sm.services, srv)
}
