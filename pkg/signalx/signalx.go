package signalx

import (
	"log"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/byteflowing/base/pkg/thread"
)

const (
	defaultCap = 10
)

type SignalHandler interface {
	Start()
	Stop()
}

type StopFunc func()

func (s StopFunc) Start() {}
func (s StopFunc) Stop() {
	s()
}

type SignalListener struct {
	mux          sync.Mutex
	handlers     []SignalHandler
	hs           map[uintptr]struct{}
	waitDuration time.Duration
}

func NewSignalListener(wait time.Duration) *SignalListener {
	return &SignalListener{
		hs:           make(map[uintptr]struct{}, defaultCap),
		handlers:     make([]SignalHandler, 0, defaultCap),
		waitDuration: wait,
	}
}

func (s *SignalListener) Add(handler SignalHandler) {
	if handler == nil {
		return
	}
	s.add(handler)
}

func (s *SignalListener) AddStopFunc(f StopFunc) {
	s.add(f)
}

func (s *SignalListener) Listen() {
	for _, handler := range s.handlers {
		h := handler
		thread.GoUnsafe(func() {
			h.Start()
		})
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()
	sig := <-sigCh
	log.Printf("[signalx] received signal: %v", sig)
	stopGroup := thread.NewRoutineGroup()
	for _, handler := range s.handlers {
		h := handler
		stopGroup.RunSafe(func() {
			h.Stop()
		})
	}
	done := make(chan struct{})
	thread.GoSafe(func() {
		stopGroup.Wait()
		close(done)
	})
	select {
	case <-done:
		log.Printf("[signalx] stopped gracefully")
	case sig2 := <-sigCh:
		log.Printf("[signalx] received second signal: %v, force kill", sig2)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
	case <-time.After(s.waitDuration):
		log.Printf("[signalx] timeout after %v secods, force kill", s.waitDuration.Seconds())
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
	}
}

func (s *SignalListener) add(handler SignalHandler) {
	ptr := reflect.ValueOf(handler).Pointer()
	s.mux.Lock()
	defer s.mux.Unlock()
	// 重复实例无需添加
	if _, ok := s.hs[ptr]; ok {
		return
	}
	s.handlers = append(s.handlers, handler)
}
