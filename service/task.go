package service

import (
	. "github.com/ahmetb/go-linq/v3"
	"github.com/rs/xid"
	"sync"
	"time"
)

var DefaultTask = NewTaskPool()

const (
	TaskTypeCopy      = "Copy"
	TaskTypeMove      = "Move"
	TaskTypeSearch    = "Search"
	TaskTypeDelete    = "Delete"
	TaskTypeUnarchive = "Unarchive"
	TaskTypeArchive   = "Archive"
	TaskStateRunning  = "Running"
	TaskStateComplete = "Complete"
	TaskStateError    = "Error"
	TaskStateAnalyze  = "Analyze"
)

type Task interface {
	Run()
	GetStatus() string
	GetId() string
	GetType() string
	GetError() error
	GetStartTime() *time.Time
	GetStopTime() *time.Time
	Interrupt()
	GetUsername() string
}
type TaskInfo struct {
	Id            string        `json:"id"`
	Type          string        `json:"type"`
	Status        string        `json:"status"`
	Error         error         `json:"error,omitempty"`
	InterruptChan chan struct{} `json:"-"`
	StartTime     *time.Time    `json:"start_time"`
	StopTime      *time.Time    `json:"stop_time"`
	Username      string        `json:"user"`
}

func (t *TaskInfo) GetStatus() string {
	return t.Status
}
func (t *TaskInfo) GetId() string {
	return t.Id
}
func (t *TaskInfo) GetType() string {
	return t.Type
}
func (t *TaskInfo) GetError() error {
	return t.Error
}
func (t *TaskInfo) GetStartTime() *time.Time {
	return t.StartTime
}
func (t *TaskInfo) GetStopTime() *time.Time {
	return t.StopTime
}
func (t *TaskInfo) GetUsername() string {
	return t.Username
}
func (t *TaskInfo) UpdateStopTime() {
	endTime := time.Now()
	t.StopTime = &endTime
}
func (t *TaskInfo) Interrupt() {
	t.InterruptChan <- struct{}{}
}

type TaskPool struct {
	Tasks []Task
	sync.RWMutex
}
type TaskQueryBuilder struct {
	Types      []string
	Status     []string
	Order      string
	OrderKey   string
	InUsername string
}

func (q *TaskQueryBuilder) WithTypes(types ...string) *TaskQueryBuilder {
	if len(types) == 0 {
		return q
	}
	if q.Types == nil {
		q.Types = types
		return q
	}
	q.Types = append(q.Types, types...)
	return q
}

func (q *TaskQueryBuilder) WithStatus(status ...string) *TaskQueryBuilder {
	if len(status) == 0 {
		return q
	}
	if q.Status == nil {
		q.Status = status
		return q
	}
	q.Status = append(q.Status, status...)
	return q
}
func (q *TaskQueryBuilder) WithOrder(key string, order string) *TaskQueryBuilder {
	q.OrderKey = key
	q.Order = order
	return q
}
func (q *TaskQueryBuilder) InUser(username string) *TaskQueryBuilder {
	q.InUsername = username
	return q
}

func (q *TaskQueryBuilder) Query() []Task {
	// default value
	if len(q.OrderKey) == 0 {
		q.OrderKey = "startTime"
	}
	if len(q.Order) == 0 {
		q.Order = "desc"
	}

	tasks := DefaultTask.GetAllTask()
	query := From(tasks)
	if q.Types != nil {
		query = query.Where(func(i interface{}) bool {
			for _, targetType := range q.Types {
				switch targetType {
				case TaskTypeCopy:
					if _, ok := i.(*CopyTask); ok {
						return true
					}
				case TaskTypeSearch:
					if _, ok := i.(*SearchFileTask); ok {
						return true
					}
				case TaskTypeDelete:
					if _, ok := i.(*DeleteFileTask); ok {
						return true
					}
				case TaskTypeUnarchive:
					if _, ok := i.(*ExtractTask); ok {
						return true
					}
				}
			}
			return false
		})
	}
	if q.Status != nil {
		query = query.Where(func(i interface{}) bool {
			for _, targetStatus := range q.Status {
				if i.(Task).GetStatus() == targetStatus {
					return true
				}
			}
			return false
		})
	}

	query = query.Sort(func(i, j interface{}) bool {
		iTask := i.(Task)
		jTask := j.(Task)
		if q.OrderKey == "startTime" {
			if q.Order == "asc" {
				return iTask.GetStartTime().Before(*jTask.GetStartTime())
			}
			if q.Order == "desc" {
				return !iTask.GetStartTime().Before(*jTask.GetStartTime())
			}
		}
		return true
	})
	query = query.Where(func(i interface{}) bool {
		return i.(Task).GetUsername() == q.InUsername
	})
	query.ToSlice(&tasks)
	return tasks
}
func NewTaskPool() *TaskPool {
	return &TaskPool{Tasks: make([]Task, 0)}
}
func (t *TaskPool) GetAllTask() []Task {
	return t.Tasks
}
func (t *TaskPool) GetTask(id string) Task {
	for _, task := range t.Tasks {
		if task.GetId() == id {
			return task
		}
	}
	return nil
}
func (t *TaskPool) StopTask(id string) {
	for _, task := range t.Tasks {
		if task.GetId() == id {
			task.Interrupt()
		}
	}
}

func (t *TaskPool) createTask(username string) TaskInfo {
	task := TaskInfo{
		Id:            xid.New().String(),
		InterruptChan: make(chan struct{}),
		Username:      username,
	}
	startTime := time.Now()
	task.StartTime = &startTime
	return task
}
