package service

import (
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"sync"
)

type ExtractInput struct {
	Input    string
	Output   string
	Password string
}
type ExtractTaskOption struct {
	OnComplete            func(id string)
	OnFileExtractComplete func(id string, output string)
}
type ExtractTask struct {
	TaskInfo
	Input      []*ExtractInput    `json:"-"`
	Option     *ExtractTaskOption `json:"-"`
	OnComplete func(id string)    `json:"-"`
	Output     *ExtractTaskOutput
	sync.Mutex
}
type ExtractTaskOutput struct {
	Complete int      `json:"complete"`
	Total    int      `json:"total"`
	Files    []string `json:"name"`
	Path     []string `json:"path"`
}

func (t *TaskPool) NewExtractTask(input []*ExtractInput, option ExtractTaskOption) *ExtractTask {
	taskInfo := t.createTask()
	taskInfo.Type = TaskTypeUnarchive
	taskInfo.Status = TaskStateRunning
	task := &ExtractTask{
		TaskInfo: taskInfo,
		Input:    input,
		Option:   &option,
	}
	o := &ExtractTaskOutput{
		Path:     []string{},
		Files:    []string{},
		Total:    len(input),
		Complete: 0,
	}
	for _, extractInput := range input {
		o.Path = append(o.Path, extractInput.Input)
		o.Files = append(o.Files, filepath.Base(extractInput.Input))
	}
	task.Output = o
	t.Lock()
	t.Tasks = append(t.Tasks, task)
	t.Unlock()
	return task
}
func (t *ExtractTask) Run() {
	t.Lock()
	t.Status = TaskStateRunning
	t.Unlock()
	for _, input := range t.Input {
		err := ExtractArchive(ExtractFileOption{
			Input:    input.Input,
			Output:   input.Output,
			Password: input.Password,
		})
		if err != nil {
			logrus.Error(err)
		}
		t.Output.Complete += 1
		if t.Option.OnFileExtractComplete != nil {
			t.Option.OnFileExtractComplete(t.Id, input.Output)
		}
	}
	t.Lock()
	t.Status = TaskStateComplete
	if t.OnComplete != nil {
		t.OnComplete(t.Id)
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
	taskInfo.Type = TaskTypeArchive
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
