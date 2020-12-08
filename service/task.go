package service

import (
	"github.com/rs/xid"
	"os"
	"sync"
	"time"
)

var DefaultTask *TaskPool = NewTaskPool()

const (
	TaskTypeCopy      = "Copy"
	TaskTypeSearch    = "Search"
	TaskStateRunning  = "Running"
	TaskStateComplete = "Complete"
	TaskStateError    = "Error"
	TaskStateAnalyze  = "Analyze"
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

func (t *TaskPool) NewCopyFileTask(src, dest string) *Task {
	task := &Task{
		Id:     xid.New().String(),
		Type:   TaskTypeCopy,
		Status: TaskStateAnalyze,
		Output: &CopyFileTaskOutput{
			Src:  src,
			Dest: dest,
		},
	}
	t.Lock()
	t.Tasks = append(t.Tasks, task)
	t.Unlock()
	go func() {
		output := task.Output.(*CopyFileTaskOutput)
		// analyze
		copyInfo, err := analyzeCopySource(src)
		t.Lock()
		if err != nil {
			t.Lock()
			task.Error = err
			task.Status = TaskStateError
			return
		}
		output.FileCount = copyInfo.FileCount
		output.TotalLength = copyInfo.TotalSize
		task.Status = TaskStateRunning
		t.Unlock()

		notifier := &CopyFileNotifier{
			CurrentFileChan:   make(chan string),
			CompleteDeltaChan: make(chan int64),
			FileCompleteChan:  make(chan string),
		}
		// update info
		go func() {
			var completeLength int64 = 0
			var completeCount int = 0
			ticker := time.NewTicker(2 * time.Second)
			var lastComplete int64 = 0
			for {
				select {
				case currentFile := <-notifier.CurrentFileChan:
					t.Lock()
					output.CurrentCopy = currentFile
					t.Unlock()
				case completeDelta := <-notifier.CompleteDeltaChan:
					t.Lock()
					completeLength += completeDelta
					output.CompleteLength = completeLength

					output.Progress = float64(completeLength) / float64(output.TotalLength)
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
					nowLength := output.CompleteLength
					t.Lock()
					output.Speed = nowLength - lastComplete
					//fmt.Printf("%s/s \n", humanize.Bytes(uint64(nowLength)-uint64(lastComplete)))
					t.Unlock()
					lastComplete = output.CompleteLength
				case <-notifier.FileCompleteChan:
					t.Lock()
					completeCount += 1
					output.Complete = completeCount
					t.Unlock()
					if completeCount == output.FileCount {
						//fmt.Println(" copy complete")
						t.Lock()
						output.CompleteLength = output.TotalLength
						output.Progress = 1
						return
					}
				}
			}
		}()
		err = Copy(src, dest, notifier)
		t.Lock()
		if err != nil {
			t.Lock()
			task.Error = err
			task.Status = TaskStateError
			return
		}
		task.Status = TaskStateComplete
		t.Unlock()
	}()
	return task
}

type CopyFileTaskOutput struct {
	TotalLength    int64   `json:"total_length"`
	FileCount      int     `json:"file_count"`
	Complete       int     `json:"complete"`
	CompleteLength int64   `json:"complete_length"`
	Src            string  `json:"src"`
	Dest           string  `json:"dest"`
	CurrentCopy    string  `json:"current_copy"`
	Progress       float64 `json:"progress"`
	Speed          int64   `json:"speed"`
}
