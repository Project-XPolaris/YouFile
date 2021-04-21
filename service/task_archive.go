package service

import (
	"github.com/mholt/archiver/v3"
	"sync"
)

type UnarchiveTask struct {
	TaskInfo
	Source     string                         `json:"source"`
	Target     string                         `json:"target"`
	OnComplete func(id string, target string) `json:"-"`
	sync.Mutex
}

func (t *TaskPool) NewUnarchiveTask(source string, target string, OnComplete func(id string, target string)) *UnarchiveTask {
	taskInfo := t.createTask()
	taskInfo.Type = TaskTypeUnarchive
	taskInfo.Status = TaskStateRunning
	task := &UnarchiveTask{
		TaskInfo:   taskInfo,
		Source:     source,
		Target:     target,
		OnComplete: OnComplete,
	}
	t.Lock()
	t.Tasks = append(t.Tasks, task)
	t.Unlock()
	return task
}
func (t *UnarchiveTask) Run() {
	t.Lock()
	t.Status = TaskStateRunning
	t.Unlock()
	err := archiver.Unarchive(t.Source, t.Target)
	if err != nil {
		t.Lock()
		t.Error = err
		t.Status = TaskStateError
		t.Unlock()
	}
	t.Lock()
	t.Status = TaskStateComplete
	if t.OnComplete != nil {
		t.OnComplete(t.Id, t.Target)
	}
	t.Unlock()

}

type ArchiveTask struct {
	TaskInfo
	Sources    []string                       `json:"sources"`
	Target     string                         `json:"target"`
	OnComplete func(id string, target string) `json:"-"`
	sync.Mutex
}

func (t *TaskPool) NewArchiveTask(source []string, target string, OnComplete func(id string, target string)) *ArchiveTask {
	taskInfo := t.createTask()
	taskInfo.Type = TaskTypeUnarchive
	taskInfo.Status = TaskStateRunning
	task := &ArchiveTask{
		TaskInfo:   taskInfo,
		Sources:    source,
		Target:     target,
		OnComplete: OnComplete,
	}
	t.Lock()
	t.Tasks = append(t.Tasks, task)
	t.Unlock()
	return task
}
func (t *ArchiveTask) Run() {
	t.Lock()
	t.Status = TaskStateRunning
	t.Unlock()
	err := archiver.Archive(t.Sources, t.Target)
	if err != nil {
		t.Lock()
		t.Error = err
		t.Status = TaskStateError
		t.Unlock()
	}
	t.Lock()
	t.Status = TaskStateComplete
	if t.OnComplete != nil {
		t.OnComplete(t.Id, t.Target)
	}
	t.Unlock()
}
