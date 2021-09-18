package service

import (
	"sync"
	"time"
)

type DeleteFileTaskOutput struct {
	FileCount     int      `json:"file_count"`
	Complete      int      `json:"complete"`
	Src           []string `json:"src"`
	Progress      float64  `json:"progress"`
	Speed         int      `json:"speed"`
	CurrentDelete string   `json:"current_delete"`
}

type DeleteFileTask struct {
	TaskInfo
	Output *DeleteFileTaskOutput
	Option *NewDeleteFileTaskOption
	sync.Mutex
}

type NewDeleteFileTaskOption struct {
	Src            []string
	OnDone         func(id string)
	OnItemComplete func(id string, src string)
	Username       string
}

func (t *TaskPool) NewDeleteFileTask(option *NewDeleteFileTaskOption) Task {
	taskInfo := t.createTask(option.Username)
	taskInfo.Type = TaskTypeDelete
	taskInfo.Status = TaskStateRunning
	task := DeleteFileTask{
		TaskInfo: taskInfo,
		Output: &DeleteFileTaskOutput{
			Src: option.Src,
		},
		Option: option,
	}
	t.Lock()
	t.Tasks = append(t.Tasks, &task)
	t.Unlock()
	return &task
}
func (t *DeleteFileTask) Run() {
	// analyze
	infos := make([]*CopyAnalyzeResult, 0)
	for _, deleteSrc := range t.Option.Src {
		copyInfo, err := analyzeSource(deleteSrc)
		if err != nil {
			t.Lock()
			t.Error = err
			t.Status = TaskStateError
			t.Unlock()
			return
		}
		infos = append(infos, copyInfo)
	}

	t.Lock()
	t.Output.FileCount = 0
	for _, info := range infos {
		t.Output.FileCount += info.FileCount
	}
	t.Status = TaskStateRunning
	t.Unlock()

	notifier := &DeleteNotifier{
		DeleteChan:     make(chan string),
		DeleteDoneChan: make(chan string),
	}
	// update info
	go func() {
		var completeCount = 0
		ticker := time.NewTicker(2 * time.Second)
		var lastComplete = 0
		for {
			select {
			case currentFile := <-notifier.DeleteChan:
				t.Lock()
				t.Output.CurrentDelete = currentFile
				t.Unlock()
			case <-ticker.C:
				nowCount := t.Output.Complete
				t.Lock()
				t.Output.Speed = nowCount - lastComplete
				t.Unlock()
				lastComplete = t.Output.Complete
			case <-t.InterruptChan:
				t.Lock()
				notifier.StopFlag = true
				t.Unlock()
			case <-notifier.DeleteDoneChan:
				t.Lock()
				completeCount += 1
				t.Output.Complete = completeCount
				t.Output.Progress = float64(completeCount) / float64(t.Output.FileCount)
				t.Unlock()
				if completeCount == t.Output.FileCount {
					//fmt.Println(" copy complete")
					t.Lock()
					t.Output.Progress = 1
					t.Unlock()
					return
				}
			}
		}
	}()
	for _, deleteSrc := range t.Output.Src {
		err := Delete(deleteSrc, notifier)
		if err != nil {
			if err == DeleteInterrupt {
				break
			}
			t.Lock()
			t.Error = err
			t.UpdateStopTime()
			t.Unlock()
			return
		}
		if t.Option.OnItemComplete != nil {
			t.Option.OnItemComplete(t.Id, deleteSrc)
		}
	}

	t.Lock()
	t.Status = TaskStateComplete
	t.UpdateStopTime()
	t.Unlock()
	if t.Option.OnDone != nil {
		t.Option.OnDone(t.Id)
	}
}
