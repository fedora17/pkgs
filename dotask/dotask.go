package dotask

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"pkgs/logger"
	"sort"
	"strings"
	"sync"
	"time"
)

var location = time.Local

type JobFunc func(id ID)

// SkipJobIfRunning 是对当前任务的一个封装
// 使当前任务只能串行 上次没执行完就跳过此次执行
func SkipJobIfRunning(j JobFunc) JobFunc {
	var n int
	var ch = make(chan struct{}, 1)
	ch <- struct{}{}

	return func(id ID) {
		select {
		case v := <-ch:
			j(id)
			ch <- v
		default:
			n++
			logger.Warnf("skipped already running task, id=%d n=%d", id, n)
		}
	}
}

// DelayJobIfRunning 是对当前任务的一个封装
// 使当前任务只能串行 等到上次执行完才执行本次
func DelayJobIfRunning(j JobFunc) JobFunc {
	var n int
	var mu sync.Mutex

	return func(id ID) {
		start := time.Now()
		mu.Lock()
		defer mu.Unlock()
		// 等待超过1m代表已被延时执行
		if since := time.Since(start); since > time.Minute {
			n++
			logger.Warnf("delayed execution task id=%d, duration=%s, n=%d", id, since.String(), n)
		}
		j(id)
	}
}

type ID uint64

type Task struct {
	ID         ID        `json:"id"`
	Times      uint64    `json:"times"`
	Limit      uint64    `json:"limit"`
	Next       time.Time `json:"next"`
	Prev       time.Time `json:"prev"`
	Expr       string    `json:"expr"`
	Title      string    `json:"title"`
	CreateTime time.Time `json:"create_time"`
	jobFunc    JobFunc
	cronExpr   *cronexpr.Expression
}

func (t *Task) validExpr() bool {
	return len(strings.SplitN(t.Expr, " ", 2)) == 2
}

// next 将根据输入时间和表达式返回下一次执行时间
// 超过执行次数限制或无需再执行 返回zero(丢弃)
func (t *Task) next(prev time.Time) time.Time {
	var zero time.Time

	if t.Limit != 0 && t.Limit <= t.Times {
		return zero
	}

	spn := strings.SplitN(t.Expr, " ", 2)
	flag, expr := spn[0], spn[1]
	switch flag {
	// like "@ 2019-12-25 20:08:00"
	case "@", "at":
		if t.Times > 0 {
			return zero
		}
		at, err := time.ParseInLocation("2006-01-02 15:04:05", expr, location)
		if err != nil {
			logger.Warnf("failed to parse t.expr=%s err=%v", t.Expr, err)
			return zero
		}
		// 下次执行时间必须在创建时间之后才执行 相等不执行 没那么巧吧？
		if at.After(prev) {
			return at
		}

	// like "every 1h30m"
	case "every":
		d, err := time.ParseDuration(expr)
		if err != nil {
			logger.Warnf("failed to parse t.expr=%s err=%v", t.Expr, err)
			return zero
		}
		return prev.Add(d)

	// like linux crontab "cron */10 * * * *"
	case "cron":
		return t.cronExpr.Next(prev)
	}

	return zero
}

func (t *Task) String() string {
	return fmt.Sprintf("[id=%d, title=%s, expr=%s, times/limit=%d/%d, create=%s, next=%s]",
		t.ID,
		t.Title,
		t.Expr,
		t.Times,
		t.Limit,
		t.CreateTime.Format("2006-01-02 15:04:05"),
		t.Next.Format("2006-01-02 15:04:05"),
	)
}

// parseExpr 提前校验expr的正确性
// 如果使用的cron将使用cronexpr解析
// 返回可能的错误
func (t *Task) parseExpr() error {
	if !t.validExpr() {
		return fmt.Errorf("failed to parse expr unknown format")
	}

	spn := strings.SplitN(t.Expr, " ", 2)
	flag, expr := spn[0], spn[1]
	switch flag {
	case "@", "at":
		_, err := time.ParseInLocation("2006-01-02 15:04:05", expr, location)
		if err != nil {
			return fmt.Errorf("failed to parse expr %v", err)
		}
	case "every":
		_, err := time.ParseDuration(expr)
		if err != nil {
			return fmt.Errorf("failed to parse expr %v", err)
		}
	case "cron":
		var err error
		t.cronExpr, err = cronexpr.Parse(expr)
		if err != nil {
			return fmt.Errorf("failed to parse expr %v", err)
		}
	default:
		return fmt.Errorf("failed to parse expr invalid flag=%s", flag)
	}

	return nil
}

type DoTask struct {
	tasks    []*Task
	add      chan *Task
	remove   chan ID
	stop     chan struct{}
	wg       sync.WaitGroup
	snapshot chan chan []Task
	nextID   ID
	mu       sync.Mutex
}

func New() *DoTask {
	return &DoTask{
		tasks:    nil,
		add:      make(chan *Task),
		remove:   make(chan ID),
		stop:     make(chan struct{}),
		wg:       sync.WaitGroup{},
		snapshot: make(chan chan []Task),
		nextID:   0,
		mu:       sync.Mutex{},
	}
}

var _default = New()

func Add(job func(id ID), expr string, title string, limit uint64) ID {
	return _default.Add(job, expr, title, limit)
}
func Remove(id ID) {
	_default.Remove(id)
}
func GetTasks() []Task {
	return _default.GetTasks()
}
func Run() {
	_default.Run()
}
func Start() {
	go _default.Run()
}

func Stop() {
	_default.Stop()
}

// Add 添加任务到任务队列 期间将解析expr
// 返回添加的任务id expr错误时id=0
func (d *DoTask) Add(job func(id ID), expr string, title string, limit uint64) ID {
	d.mu.Lock()
	defer d.mu.Unlock()

	task := &Task{
		Expr:       expr,
		Title:      title,
		Limit:      limit,
		CreateTime: d.now(),
		jobFunc:    job,
	}
	err := task.parseExpr()
	if err != nil {
		logger.Warn(err)
		return ID(0)
	}
	d.nextID++
	task.ID = d.nextID
	task.Next = task.next(task.CreateTime)

	d.add <- task

	return task.ID
}

func (d *DoTask) Remove(id ID) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.remove <- id
}

func (d *DoTask) removeTask(id ID) *Task {
	var task *Task
	var tasks []*Task
	for _, t := range d.tasks {
		if t.ID != id {
			tasks = append(tasks, t)
		} else {
			task = t
		}
	}
	d.tasks = tasks

	return task
}

func (d *DoTask) GetTasks() []Task {
	d.mu.Lock()
	defer d.mu.Unlock()

	replyChan := make(chan []Task, 1)
	d.snapshot <- replyChan
	return <-replyChan
}

func (d *DoTask) taskSnapshot() []Task {
	var tasks = make([]Task, len(d.tasks))
	for i, t := range d.tasks {
		tasks[i] = *t
	}
	return tasks
}

func (d *DoTask) now() time.Time {
	return time.Now().In(location)
}

func (d *DoTask) startTask(t *Task) {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		t.jobFunc(t.ID)
		logger.Info("schedule done ", t)
	}()
}

// 按时间升序 zero代表执行过 丢最后面
func (d *DoTask) sortByTime() {
	sort.SliceStable(d.tasks, func(i, j int) bool {
		if d.tasks[i].Next.IsZero() {
			return false
		}
		if d.tasks[j].Next.IsZero() {
			return true
		}
		return d.tasks[i].Next.Before(d.tasks[j].Next)
	})
}

// Stop 停止调度新任务并退出
// 若还有在运行的任务 等待1m
func (d *DoTask) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.stop <- struct{}{}
	<-d.stop

	waitChan := make(chan struct{}, 1)
	defer close(waitChan)
	go func() {
		d.wg.Wait()
		waitChan <- struct{}{}
	}()

	waitTimer := time.NewTimer(time.Minute)
	defer waitTimer.Stop()

	select {
	case <-waitTimer.C:
		logger.Warn("waiting for task to finish, timed out!")
	case <-waitChan:
	}

	for _, task := range d.tasks {
		if task.Next.Before(d.now()) {
			continue
		}

		if task.Next.IsZero() {
			break
		}
		logger.Info("abort active task ", task)
	}
}

func (d *DoTask) Run() {
	logger.Info("dotask is started!")
	now := d.now()

	for {
		d.sortByTime()

		var timer *time.Timer
		if len(d.tasks) == 0 || d.tasks[0].Next.IsZero() {
			logger.Debug("no task can be schedule")
			timer = time.NewTimer(100000 * time.Hour)
		} else {
			logger.Debug("next schedule at ", d.tasks[0].Next.String())
			timer = time.NewTimer(d.tasks[0].Next.Sub(now))
		}

		for {
			select {
			case now = <-timer.C:
				now = now.In(location)
				for _, task := range d.tasks {
					if task.Next.After(now) || task.Next.IsZero() {
						break
					}
					task.Times++
					d.startTask(task)
					task.Prev = task.Next
					task.Next = task.next(task.Prev)
					if task.Next.IsZero() {
						logger.Info("last scheduled of the task ", task)
					}
					logger.Info("schedule start ", task)
				}

			case task := <-d.add:
				timer.Stop()
				now = d.now()
				d.tasks = append(d.tasks, task)
				logger.Info("add a new task ", task)

			case id := <-d.remove:
				timer.Stop()
				now = d.now()
				task := d.removeTask(id)
				if task == nil {
					logger.Infof("failed to remove task, not found! id=%d", id)
				} else {
					logger.Info("removed task ", task)
				}

			case replyChan := <-d.snapshot:
				replyChan <- d.taskSnapshot()
				continue

			case <-d.stop:
				timer.Stop()
				logger.Info("schedule stopped!")
				d.stop <- struct{}{}
				return
			}

			break
		}
	}
}
