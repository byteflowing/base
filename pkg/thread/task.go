package thread

import (
	"sync"
	"sync/atomic"
)

type TaskRunner struct {
	limits chan struct{}
	wg     sync.WaitGroup
	closed atomic.Bool
}

func NewTaskRunner(limits int) *TaskRunner {
	return &TaskRunner{limits: make(chan struct{}, limits)}
}

func (t *TaskRunner) Schedule(task func()) {
	if t.closed.Load() {
		return
	}
	t.wg.Add(1)
	t.limits <- struct{}{}
	GoUnsafe(func() {
		defer func() {
			<-t.limits
			t.wg.Done()
		}()
		task()
	})
}

func (t *TaskRunner) Wait() {
	if t.closed.Swap(true) {
		return
	}
	t.wg.Wait()
}
