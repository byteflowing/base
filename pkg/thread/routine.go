package thread

import (
	"log"
	"runtime/debug"
	"sync"
)

type RoutineGroup struct {
	wg sync.WaitGroup
}

func NewRoutineGroup() *RoutineGroup {
	return &RoutineGroup{}
}

func (g *RoutineGroup) RunSafe(fn func()) {
	g.wg.Add(1)
	GoSafe(func() {
		defer g.wg.Done()
		fn()
	})
}

func (g *RoutineGroup) Run(fn func()) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()
		fn()
	}()
}

func (g *RoutineGroup) Wait() {
	g.wg.Wait()
}

func GoSafe(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[GoSafe] panic: %v\n%s", r, debug.Stack())
			}
		}()
		fn()
	}()
}

func GoUnsafe(fn func()) {
	go fn()
}
