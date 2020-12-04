package service

import (
	"github.com/rs/xid"
	"os"
	"sync"
)

var DefaultTask *TaskPool = NewTaskPool()

const (
	TaskTypeCopy      = "Copy"
	TaskTypeSearch    = "Search"
	TaskStateRunning  = "Running"
	TaskStateComplete = "Complete"
	TaskStateError    = "Error"
)

type Task struct {
	Id     string      `json:"id"`
	Type   string      `json:"type"`
	Status string      `json:"status"`
	Output interface{} `json:"output,omitempty"`
	Error  error       `json:"error,omitempty"`
}
type TaskPool struct {
	Tasks []*Task
	sync.RWMutex
}

func NewTaskPool() *TaskPool {
	return &TaskPool{Tasks: make([]*Task, 0)}
}
func (t *TaskPool) GetTask(id string) *Task {
	for _, task := range t.Tasks {
		if task.Id == id {
			return task
		}
	}
	return nil
}

type ScanFileOutput struct {
	Parent string
	Files  []os.FileInfo
}

func (t *TaskPool) NewScanFileTask(src string, key string) *Task {
	task := &Task{
		Id:     xid.New().String(),
		Type:   TaskTypeSearch,
		Status: TaskStateRunning,
	}
	t.Lock()
	t.Tasks = append(t.Tasks, task)
	t.Unlock()
	go func() {
		info, err := SearchFile(src, key)
		if err != nil {
			t.Lock()
			task.Error = err
			task.Status = TaskStateError
			t.Unlock()
		}
		t.Lock()
		task.Output = ScanFileOutput{
			Parent: src,
			Files:  info,
		}
		task.Status = TaskStateComplete
		t.Unlock()
	}()
	return task
}
