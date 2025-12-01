package thread

type WorkGroup struct {
	job    func()
	limits int
}

func NewWorkGroup(job func(), limits int) *WorkGroup {
	return &WorkGroup{
		job:    job,
		limits: limits,
	}
}

func (w *WorkGroup) Run() {
	group := NewRoutineGroup()
	for i := 0; i < w.limits; i++ {
		group.Run(w.job)
	}
	group.Wait()
}
