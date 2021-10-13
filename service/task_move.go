package service

import (
	"sync"
	"time"
	"youfile/util"
)

type NewMoveTaskOption struct {
	Options     []*MoveOption
	OnDone      func(task *MoveTask)
	OnError     func(task *MoveTask)
	Username    string `json:"username"`
	OnDuplicate string `json:"onDuplicate"`
	DisplayPath map[string]string
}
type MoveOption struct {
	Src        string          `json:"src"`
	Dest       string          `json:"dest"`
	OnComplete func(id string) `json:"-"`
}
type MoveTask struct {
	TaskInfo
	Output *MoveFileTaskOutput
	sync.Mutex
	Option *NewMoveTaskOption
}

type MoveFileTaskOutput struct {
	TotalLength    int64         `json:"total_length"`
	FileCount      int           `json:"file_count"`
	Complete       int           `json:"complete"`
	CompleteLength int64         `json:"complete_length"`
	List           []*MoveOption `json:"list"`
	CurrentMove    string        `json:"current_copy"`
	Progress       float64       `json:"progress"`
	Speed          int64         `json:"speed"`
}

func (t *MoveTask) AbortError(err error) {
	t.Lock()
	t.Error = err
	t.Status = TaskStateError
	t.UpdateStopTime()
	t.Unlock()
	if t.Option.OnError != nil {
		t.Option.OnError(t)
	}
}
func (t *TaskPool) NewMoveTask(option *NewMoveTaskOption) Task {
	taskInfo := t.createTask(option.Username)
	taskInfo.Type = TaskTypeMove
	taskInfo.Status = TaskStateRunning
	task := MoveTask{
		TaskInfo: taskInfo,
		Output: &MoveFileTaskOutput{
			List: option.Options,
		},
		Option: option,
	}
	t.Lock()
	t.Tasks = append(t.Tasks, &task)
	t.Unlock()
	return &task
}
func (t *MoveTask) Run() {
	// analyze
	infos := make([]*CopyAnalyzeResult, 0)
	for _, option := range t.Option.Options {
		copyInfo, err := analyzeSource(option.Src)
		if err != nil {
			t.AbortError(err)
			return
		}
		infos = append(infos, copyInfo)
	}

	t.Lock()
	t.Output.FileCount = 0
	t.Output.TotalLength = 0
	for _, info := range infos {
		t.Output.FileCount += info.FileCount
		t.Output.TotalLength += info.TotalSize
	}
	t.Status = TaskStateRunning
	t.Unlock()

	notifier := &MoveFileNotifier{
		CurrentFileChan:   make(chan string),
		CompleteDeltaChan: make(chan int64),
		FileCompleteChan:  make(chan string),
		StopChan:          make(chan struct{}, 1),
	}
	// update info
	go func() {
		var completeLength int64 = 0
		var completeCount = 0
		ticker := time.NewTicker(2 * time.Second)
		var lastComplete int64 = 0
		for {
			select {
			case currentFile := <-notifier.CurrentFileChan:
				t.Lock()
				t.Output.CurrentMove = currentFile
				t.Unlock()
			case completeDelta := <-notifier.CompleteDeltaChan:
				t.Lock()
				completeLength += completeDelta
				t.Output.CompleteLength = completeLength

				t.Output.Progress = float64(completeLength) / float64(t.Output.TotalLength)
				//fmt.Printf("current file: %s count: %d/%d lenght: %d/%d   %.2f \n",
				//	filepath.Base(output.CurrentCopy),
				//	completeCount,
				//	output.FileCount,
				//	completeLength,
				//	output.TotalLength,
				//	(float64(completeLength)/float64(output.TotalLength))*100,
				//)
				t.Unlock()

			case <-ticker.C:
				nowLength := t.Output.CompleteLength
				t.Lock()
				t.Output.Speed = nowLength - lastComplete
				//fmt.Printf("%s/s \n", humanize.Bytes(uint64(nowLength)-uint64(lastComplete)))
				t.Unlock()
				lastComplete = t.Output.CompleteLength
			case <-notifier.FileCompleteChan:
				t.Lock()
				completeCount += 1
				t.Output.Complete = completeCount
				t.Unlock()
				if completeCount == t.Output.FileCount {
					//fmt.Println(" copy complete")
					t.Lock()
					t.Output.CompleteLength = t.Output.TotalLength
					t.Output.Progress = 1
					t.Unlock()
					return
				}
			case <-t.InterruptChan:
				notifier.StopFlag = true
				notifier.StopChan <- struct{}{}
			}
		}
	}()
	for _, option := range t.Option.Options {
		err := Move(option.Src, option.Dest, notifier, t.Option.OnDuplicate)
		if err == util.CopyInterrupt {
			break
		}
		if err != nil {
			t.AbortError(err)
			return
		}
		if option.OnComplete != nil {
			option.OnComplete(t.Id)
		}
	}
	t.Lock()
	t.Status = TaskStateComplete
	t.UpdateStopTime()
	t.Unlock()
	if t.Option.OnDone != nil {
		t.Option.OnDone(t)
	}
}
