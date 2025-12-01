package cron

import (
	"time"

	"github.com/robfig/cron/v3"
)

type (
	Entry      = cron.Entry
	EntryID    = cron.EntryID
	Job        = cron.Job
	Schedule   = cron.Schedule
	Option     = cron.Option
	Parser     = cron.Parser
	Logger     = cron.Logger
	JobWrapper = cron.JobWrapper
	FunJob     = cron.FuncJob
)

var (
	DefaultLogger       = cron.DefaultLogger
	NewParser           = cron.NewParser
	WithLogger          = cron.WithLogger
	WithLocation        = cron.WithLocation
	WithSeconds         = cron.WithSeconds
	WithChain           = cron.WithChain
	SkipIfStillRunning  = cron.SkipIfStillRunning
	DelayIfStillRunning = cron.DelayIfStillRunning
	Recover             = cron.Recover
	Every               = cron.Every
)

type Cron struct {
	cron *cron.Cron
}

// New 创建一个新的 cron 调度器（scheduler）
//
//	选项：
//	WithLocation 可选 指定调度器的时区，建议使用 time.Local
//	WithLogger 可选 指定日志对象（实现Logger接口）
//	WithChain 指定调度任务执行过程中的错误处理链
//
// 我们可以在初始化 cron 时，为任务定义一系列的装饰器，这样当有新的任务注册进来，就能自动为其附加装饰器的所有功能
//
//	cron.Recover：恢复任务执行过程中产生的 panic，不要让 cron 调度器退出
//	cron.DelayIfStillRunning：如果上一次任务还未完成，那么延迟此次任务的执行时间，只有上一次任务执行完成后，才会执行下一次任务
//	cron.SkipIfStillRunning：如果上一次任务还未完成，那么放弃此次此次任务的执行
//	cron.WithChain 会将所有装饰器串联起来，使其成为一条任务链，比如 cron.WithChain(m1, m2)，那么最终任务执行时会这样调用：m1(m2(job))
//	c := cron.New(
//
//	cron.WithSeconds(),      // 增加秒解析
//	cron.WithLogger(logger), // 自定义日志
//	cron.WithChain( // chain 是顺序敏感的
//	    cron.SkipIfStillRunning(logger), // 如果作业仍在运行，则跳过此次运行
//	    cron.Recover(logger),            // 恢复 panic
//	),
//
//	)
//
// 本cron是github.com/robfig/cron/v3的封装,在初始化时已经加上了WithSeconds()选项
//
//	spce为六字段 "秒 分 时 日 月 周"
//	秒：0-59 每分钟的哪一秒触发
//	分：0-59 每小时的哪一分钟触发
//	时：0-23 每天的哪一小时触发
//	日：1-31 每月的哪一天触发
//	月：1-12 每年的哪一月触发
//	周：0-6 每周的哪一天触发（0=星期日）
//
// 通配符：
//
//	"-" :表示“每一个” "* * * * * *"：每秒执行一次 "0 * * * * *"：每分钟第 0 秒执行（即每分钟执行一次） "0 0 * * * *"：每小时整点执行一次
//	"/" : 表示“步长”，每隔多少执行一次 "*/5 * * * * *"：每 5 秒执行一次（0,5,10,…,55）"0 */10 * * * *"：每 10 分钟执行一次（0,10,20,…,50）
//	"," : 表示“多个具体值” "0 0 9,15 * * *"：每天 9 点和 15 点各执行一次 "0 0 * * * 1,3,5"：每周一、三、五执行一次 "0 30 8,20 * * *"：每天 8:30 和 20:30 执行
//	"-" : 表示“范围” "0 0 9-17/2 * * *"：每天 9 点到 17 点之间，每 2 小时执行一次（9,11,13,15,17）"0 0 9-17 * * *"：每天 9 点到 17 点整点执行（共 9 次）"0 0 * * 1-5 *"：每月的工作日（周一到周五）执行
//
// 简写形式:
//
//	@every 5s 每 5 秒执行一次
//	@hourly 没小时执行一次
//	@daily 每天 0 点执行一次
//	@weekly 每周日 0 点执行
//	@monthly 每月 1 日 0 点
//	@yearly 每年 1 月 1 日 0 点
func New(opts ...Option) *Cron {
	return &Cron{
		cron: cron.New(opts...),
	}
}

// AddFunc 适合：逻辑短、无状态、无需依赖注入
func (c *Cron) AddFunc(spec string, cmd func()) (EntryID, error) {
	return c.cron.AddFunc(spec, cmd)
}

// AddJob 适合：需要传入结构体、带状态、可复用、易测试的任务
func (c *Cron) AddJob(spec string, job Job) (EntryID, error) {
	return c.cron.AddJob(spec, job)
}

// Schedule
//
//	是一个实现了 Schedule 接口的对象
//
// job 是要执行的任务
// 常见用法：使用 cron.Every(duration) 实现周期性任务:
//
//		c := cron.New(cron.WithSeconds())
//
//	   // 每 10 秒执行一次任务
//	   schedule := cron.Every(10 * time.Second)
//
//	   c.Schedule(schedule, cron.FuncJob(func() {
//	       fmt.Println("每 10 秒执行一次：", time.Now())
//	   }))
//
// 自定义 Schedule 示例：每天中午 12 点
//
//	type NoonSchedule struct{}
//
//	func (NoonSchedule) Next(t time.Time) time.Time {
//	   next := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
//	   if !next.After(t) {
//	       next = next.Save(24 * time.Hour)
//	   }
//	   return next
//	}
//
//	func main() {
//	   c := cron.New()
//	   c.Schedule(NoonSchedule{}, cron.FuncJob(func() {
//	       fmt.Println("每天中午执行：", time.Now())
//	   }))
//	   c.Start()
//
//	   select {}
//	}
func (c *Cron) Schedule(schedule Schedule, job Job) EntryID {
	return c.cron.Schedule(schedule, job)
}

func (c *Cron) Entries() []Entry {
	return c.cron.Entries()
}

func (c *Cron) GetEntry(id EntryID) Entry {
	return c.cron.Entry(id)
}

func (c *Cron) Remove(id EntryID) {
	c.cron.Remove(id)
}

func (c *Cron) Location() *time.Location {
	return c.cron.Location()
}

// ParseStandard 依据标准的cron表达式返回一个Schedule
// 比如讲spec写入配置文件，使用Parse解析配置文件中的定时任务配置
//
//		spec := "*/10 * * * * *" // 配置里读出来的
//		schedule, _ := cron.ParseStandard(spec)
//		c.AddJob(schedule, cron.FuncJob(func() {
//	 fmt.Println("每 10 秒执行一次")
//		}))
//		c.Start()
//
// cron标准表达式：https://en.wikipedia.org/wiki/Cron
func (c *Cron) ParseStandard(spec string) (Schedule, error) {
	return cron.ParseStandard(spec)
}

// Start 新建一个goroutines中运行（非阻塞）
func (c *Cron) Start() (err error) {
	c.cron.Start()
	return
}

// Run 在当前goroutine中运行 （阻塞）
func (c *Cron) Run() {
	c.cron.Run()
}

func (c *Cron) Stop() {
	ctx := c.cron.Stop()
	<-ctx.Done()
}
