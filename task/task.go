package task

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// err
var (
	ErrMaxRetries = fmt.Errorf("exceeded maximum retries")
)

// Job job
type Job interface {
	Run()
}

// JobWrap job wrap
type JobWrap func()

// Run run
func (f JobWrap) Run() { f() }

// Schedule  schedule
type Schedule interface {
	// Next returns the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}

// DurationSchedule duration schedule
type DurationSchedule struct {
	Interval time.Duration // 执行间隔，例如 200ms、500ms
}

// Next 实现 Schedule 接口的 Next 方法
func (ds DurationSchedule) Next(t time.Time) time.Time {
	return t.Add(ds.Interval)
}

// Every 一个便捷的构造函数，可设置最小间隔限制
func Every(duration time.Duration) DurationSchedule {
	return DurationSchedule{Interval: duration}
}

// Tasker task
type Tasker struct {
	id        int64
	cron      *cron.Cron
	tasks     map[int64]cron.EntryID
	tasksRW   sync.RWMutex
	running   sync.Map
	runningRW sync.Mutex
	opt       Options
	idMux     sync.Mutex
}

var (
	_task *Tasker
	once  sync.Once
)

// Init init
func Init(opts ...Option) *Tasker {
	once.Do(func() {
		_task = New(opts...)
	})
	return _task
}

// New new
func New(opts ...Option) *Tasker {
	opt := defaultOptions()
	for _, o := range opts {
		o(&opt)
	}
	return &Tasker{
		cron:    cron.New(cron.WithSeconds()),
		tasks:   make(map[int64]cron.EntryID),
		running: sync.Map{},
		opt:     opt,
	}
}

func (t *Tasker) checkID(id int64) error {
	_, ok := t.tasks[id]
	if ok {
		return fmt.Errorf("the task %v has been registered", id)
	}
	return nil
}

func (t *Tasker) incr() int64 {
	t.idMux.Lock()
	defer t.idMux.Unlock()
	t.id++
	return t.id
}

// ScheduleFunc schedule func
func (t *Tasker) ScheduleFunc(schedule Schedule, handler func()) (int64, error) {
	for i := 0; i < t.opt.MaxRetries; i++ {
		id := t.incr()
		if err := t.checkID(id); err != nil {
			continue
		}
		return t.ScheduleFunc3(id, schedule, handler)
	}
	return 0, ErrMaxRetries
}

// ScheduleFunc3 schedule func with id
func (t *Tasker) ScheduleFunc3(id int64, schedule Schedule, handler func()) (int64, error) {
	t.tasksRW.Lock()
	defer t.tasksRW.Unlock()

	if err := t.checkID(id); err != nil {
		return id, err
	}
	entryID := t.cron.Schedule(schedule, JobWrap(func() {
		// 上一个在执行，本次就不能被执行
		if t.IsRunning(id) {
			return
		}
		t.running.Store(id, true)
		defer func() {
			t.running.Delete(id)
		}()
		handler()
	}))
	t.tasks[id] = entryID
	return id, nil
}

// ScheduleJob schedule job
func (t *Tasker) ScheduleJob(schedule Schedule, job Job) (int64, error) {
	for i := 0; i < t.opt.MaxRetries; i++ {
		id := t.incr()
		if err := t.checkID(id); err != nil {
			continue
		}
		return t.ScheduleJob3(id, schedule, job)
	}
	return 0, ErrMaxRetries
}

// ScheduleJob3 schedule job with id
func (t *Tasker) ScheduleJob3(id int64, schedule Schedule, job Job) (int64, error) {
	t.tasksRW.Lock()
	defer t.tasksRW.Unlock()

	if err := t.checkID(id); err != nil {
		return id, err
	}

	entryID := t.cron.Schedule(schedule, JobWrap(func() {
		// 上一个在执行，本次就不能被执行
		if t.IsRunning(id) {
			return
		}
		t.running.Store(id, true)
		defer func() {
			t.running.Delete(id)
		}()
		job.Run()
	}))
	t.tasks[id] = entryID
	return id, nil
}

// AddJob add job
func (t *Tasker) AddJob(spec string, job Job) (int64, error) {
	for i := 0; i < t.opt.MaxRetries; i++ {
		id := t.incr()
		if err := t.checkID(id); err != nil {
			continue
		}
		return t.AddJob3(id, spec, job)
	}
	return 0, ErrMaxRetries
}

// AddJob3 add id job
func (t *Tasker) AddJob3(id int64, spec string, job Job) (int64, error) {
	t.tasksRW.Lock()
	defer t.tasksRW.Unlock()

	if err := t.checkID(id); err != nil {
		return id, err
	}
	entryID, err := t.cron.AddJob(spec, cron.FuncJob(func() {
		if t.IsRunning(id) {
			return
		}
		t.running.Store(id, true)
		defer func() {
			t.running.Delete(id)
		}()
		job.Run()
	}))
	if err != nil {
		return id, err
	}
	t.tasks[id] = entryID
	return id, nil
}

// RestartJob restart job
func (t *Tasker) RestartJob(id int64, spec string, job Job) (err error) {
	t.Remove(id)
	_, err = t.AddJob3(id, spec, job)
	return
}

// AddFunc add func
func (t *Tasker) AddFunc(spec string, handler func()) (int64, error) {
	for i := 0; i < t.opt.MaxRetries; i++ {
		id := t.incr()
		if err := t.checkID(id); err != nil {
			continue
		}
		return t.AddFunc3(id, spec, handler)
	}
	return 0, ErrMaxRetries
}

// AddFunc3 add func
func (t *Tasker) AddFunc3(id int64, spec string, handler func()) (int64, error) {
	t.tasksRW.Lock()
	defer t.tasksRW.Unlock()

	if err := t.checkID(id); err != nil {
		return id, err
	}
	entryID, err := t.cron.AddFunc(spec, func() {
		// 上一个在执行，本次就不能被执行
		if t.IsRunning(id) {
			return
		}
		t.running.Store(id, true)
		defer func() {
			t.running.Delete(id)
		}()
		handler()
	})
	if err != nil {
		return id, err
	}
	t.tasks[id] = entryID
	return id, nil
}

// RestartFunc restart func
func (t *Tasker) RestartFunc(id int64, spec string, handler func()) (err error) {
	t.Remove(id)
	_, err = t.AddFunc3(id, spec, handler)
	return
}

// IsExists is exists
func (t *Tasker) IsExists(id int64) (ok bool) {
	t.tasksRW.RLock()
	defer t.tasksRW.RUnlock()

	_, ok = t.tasks[id]
	return
}

// IsRunning is running
func (t *Tasker) IsRunning(id int64) bool {
	_, ok := t.running.Load(id)
	return ok
}

// Remove remove
func (t *Tasker) Remove(id int64) {
	t.tasksRW.Lock()
	defer t.tasksRW.Unlock()

	if entryID, ok := t.tasks[id]; ok {
		t.cron.Remove(entryID)
		t.running.Delete(id)
		delete(t.tasks, id)
	}
}

// Start start
func (t *Tasker) Start() {
	t.cron.Start()
}

// Stop stop
func (t *Tasker) Stop() {
	t.cron.Stop()
}
