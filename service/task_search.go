package service

import (
	"path/filepath"
	"sync"
	"time"
)

type SearchFileOutput struct {
	Files []TargetFile
}
type NewSearchTaskOption struct {
	Src       string
	Key       string
	Limit     int
	OnDone    func(id string)
	OnHit     func(id string, path string, name string, itemType string)
	PathTrans string
	Username  string
}

type SearchFileTask struct {
	TaskInfo
	Output *SearchFileOutput
	sync.Mutex
	Option *NewSearchTaskOption
}

func (t *TaskPool) NewSearchFileTask(option *NewSearchTaskOption) Task {
	taskInfo := t.createTask(option.Username)
	taskInfo.Type = TaskTypeSearch
	taskInfo.Status = TaskStateRunning
	task := SearchFileTask{
		TaskInfo: taskInfo,
		Output: &SearchFileOutput{
			Files: make([]TargetFile, 0),
		},
		Option: option,
	}

	t.Lock()
	t.Tasks = append(t.Tasks, &task)
	t.Unlock()
	return &task
}
func (t *SearchFileTask) Run() {
	notifier := &SearchFileNotifier{
		HitChan: make(chan TargetFile),
	}
	doneSearchChan := make(chan struct{})
	// watcher
	go func() {
		for {
			select {
			case file := <-notifier.HitChan:
				t.Lock()
				file.PathTrans = t.Option.PathTrans
				t.Output.Files = append(t.Output.Files, file)
				if t.Option.OnHit != nil {
					fileType := "Directory"
					if !file.Info.IsDir() {
						fileType = "File"
					}
					t.Option.OnHit(t.Id, file.Path, filepath.Base(file.Path), fileType)
				}
				t.Unlock()
			case <-t.InterruptChan:
				t.Lock()
				notifier.StopFlag = true
				t.Unlock()
			case <-doneSearchChan:
				return
			}
		}
	}()
	_, err := SearchFile(t.Option.Src, t.Option.Key, notifier, t.Option.Limit)
	doneSearchChan <- struct{}{}
	t.Lock()
	if err != nil {
		t.Error = err
		t.Status = TaskStateError
	} else {
		t.Status = TaskStateComplete
	}
	endTime := time.Now()
	t.StopTime = &endTime
	t.Unlock()
	if t.Option.OnDone != nil {
		t.Option.OnDone(t.Id)
	}
}
