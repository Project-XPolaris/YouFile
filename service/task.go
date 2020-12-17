package service

import (
	"github.com/rs/xid"
	"sync"
	"time"
	"youfile/util"
)

var DefaultTask *TaskPool = NewTaskPool()

const (
	TaskTypeCopy      = "Copy"
	TaskTypeSearch    = "Search"
	TaskTypeDelete    = "Delete"
	TaskStateRunning  = "Running"
	TaskStateComplete = "Complete"
	TaskStateError    = "Error"
	TaskStateAnalyze  = "Analyze"
)

type Task struct {
	Id            string        `json:"id"`
	Type          string        `json:"type"`
	Status        string        `json:"status"`
	Output        interface{}   `json:"output,omitempty"`
	Error         error         `json:"error,omitempty"`
	InterruptChan chan struct{} `json:"-"`
}
type TaskPool struct {
	Tasks []*Task
	sync.RWMutex
}
type TaskQueryBuilder struct {
	Types  []string
	Status []string
}

func (q *TaskQueryBuilder) WithTypes(types ...string) {
	if q.Types == nil {
		q.Types = make([]string, 0)
	}
	q.Types = append(q.Types, types...)
}

func (q *TaskQueryBuilder) WithStatus(status ...string) {
	if q.Status == nil {
		q.Status = make([]string, 0)
	}
	q.Status = append(q.Status, status...)
}

func (q *TaskQueryBuilder) Query() []*Task {
	tasks := DefaultTask.GetAllTask()
	for _, task := range tasks {
		result := make([]*Task, 0)
		if q.Types != nil {
			for _, targetType := range q.Types {
				if task.Type == targetType {
					result = append(result, task)
				}
			}
		}
		tasks = result
	}
	for _, task := range tasks {
		result := make([]*Task, 0)
		if q.Status != nil {
			for _, targetStatus := range q.Status {
				if task.Status == targetStatus {
					result = append(result, task)
				}
			}
		}
		tasks = result
	}
	return tasks
}
func NewTaskPool() *TaskPool {
	return &TaskPool{Tasks: make([]*Task, 0)}
}
func (t *TaskPool) GetAllTask() []*Task {
	return t.Tasks
}
func (t *TaskPool) GetTask(id string) *Task {
	for _, task := range t.Tasks {
		if task.Id == id {
			return task
		}
	}
	return nil
}
func (t *TaskPool) StopTask(id string) {
	for _, task := range t.Tasks {
		if task.Id == id {
			task.InterruptChan <- struct{}{}
		}
	}
}

type SearchFileOutput struct {
	Files []TargetFile
}

func (t *TaskPool) NewSearchFileTask(src string, key string, limit int) *Task {
	task := &Task{
		Id:     xid.New().String(),
		Type:   TaskTypeSearch,
		Status: TaskStateRunning,
		Output: &SearchFileOutput{
			Files: make([]TargetFile, 0),
		},
		InterruptChan: make(chan struct{}),
	}
	t.Lock()
	t.Tasks = append(t.Tasks, task)
	t.Unlock()
	go func() {
		output := task.Output.(*SearchFileOutput)
		notifier := &SearchFileNotifier{
			HitChan: make(chan TargetFile),
		}
		doneSearchChan := make(chan struct{})
		go func() {
			for {
				select {
				case file := <-notifier.HitChan:
					t.Lock()
					output.Files = append(output.Files, file)
					t.Unlock()
				case <-task.InterruptChan:
					t.Lock()
					notifier.StopFlag = true
					t.Unlock()
				case <-doneSearchChan:
					//fmt.Println("task complete")
					return
				}
			}
		}()
		_, err := SearchFile(src, key, notifier, limit)
		doneSearchChan <- struct{}{}
		t.Lock()
		if err != nil {
			task.Error = err
			task.Status = TaskStateError
		}
		task.Status = TaskStateComplete
		t.Unlock()
	}()
	return task
}

type CopyOption struct {
	Src  string `json:"src"`
	Dest string `json:"dest"`
}

func (t *TaskPool) NewCopyFileTask(options []*CopyOption) *Task {
	task := &Task{
		Id:     xid.New().String(),
		Type:   TaskTypeCopy,
		Status: TaskStateAnalyze,
		Output: &CopyFileTaskOutput{
			List: options,
		},
		InterruptChan: make(chan struct{}),
	}
	t.Lock()
	t.Tasks = append(t.Tasks, task)
	t.Unlock()
	go func() {
		output := task.Output.(*CopyFileTaskOutput)
		// analyze
		infos := make([]*CopyAnalyzeResult, 0)
		for _, option := range options {
			copyInfo, err := analyzeSource(option.Src)
			if err != nil {
				t.Lock()
				task.Error = err
				task.Status = TaskStateError
				t.Unlock()
				return
			}
			infos = append(infos, copyInfo)
		}

		t.Lock()
		output.FileCount = 0
		output.TotalLength = 0
		for _, info := range infos {
			output.FileCount += info.FileCount
			output.TotalLength += info.TotalSize
		}
		task.Status = TaskStateRunning
		t.Unlock()

		notifier := &CopyFileNotifier{
			CurrentFileChan:   make(chan string),
			CompleteDeltaChan: make(chan int64),
			FileCompleteChan:  make(chan string),
			StopChan:          make(chan struct{}, 1),
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
						t.Unlock()
						return
					}
				case <-task.InterruptChan:
					notifier.StopFlag = true
					notifier.StopChan <- struct{}{}
				}
			}
		}()
		for _, option := range options {
			err := Copy(option.Src, option.Dest, notifier)
			if err == util.CopyInterrupt {
				break
			}
			if err != nil {
				t.Lock()
				task.Error = err
				task.Status = TaskStateError
				t.Unlock()
				return
			}
		}

		t.Lock()
		task.Status = TaskStateComplete
		t.Unlock()
	}()
	return task
}

type CopyFileTaskOutput struct {
	TotalLength    int64         `json:"total_length"`
	FileCount      int           `json:"file_count"`
	Complete       int           `json:"complete"`
	CompleteLength int64         `json:"complete_length"`
	List           []*CopyOption `json:"list"`
	CurrentCopy    string        `json:"current_copy"`
	Progress       float64       `json:"progress"`
	Speed          int64         `json:"speed"`
}

type DeleteFileTaskOutput struct {
	FileCount     int     `json:"file_count"`
	Complete      int     `json:"complete"`
	Src           string  `json:"src"`
	Progress      float64 `json:"progress"`
	Speed         int     `json:"speed"`
	CurrentDelete string  `json:"current_delete"`
}

func (t *TaskPool) NewDeleteFileTask(src string) *Task {
	task := &Task{
		Id:     xid.New().String(),
		Type:   TaskTypeDelete,
		Status: TaskStateAnalyze,
		Output: &DeleteFileTaskOutput{
			Src: src,
		},
		InterruptChan: make(chan struct{}),
	}
	t.Lock()
	t.Tasks = append(t.Tasks, task)
	t.Unlock()
	go func() {
		output := task.Output.(*DeleteFileTaskOutput)
		// analyze
		copyInfo, err := analyzeSource(src)
		t.Lock()
		if err != nil {
			t.Lock()
			task.Error = err
			task.Status = TaskStateError
			return
		}
		output.FileCount = copyInfo.FileCount
		task.Status = TaskStateRunning
		t.Unlock()

		notifier := &DeleteNotifier{
			DeleteChan:     make(chan string),
			DeleteDoneChan: make(chan string),
		}
		// update info
		go func() {
			var completeCount int = 0
			ticker := time.NewTicker(2 * time.Second)
			var lastComplete int = 0
			for {
				select {
				case currentFile := <-notifier.DeleteChan:
					t.Lock()
					output.CurrentDelete = currentFile
					t.Unlock()
				case <-ticker.C:
					nowCount := output.Complete
					t.Lock()
					output.Speed = nowCount - lastComplete
					t.Unlock()
					lastComplete = output.Complete
				case <-task.InterruptChan:
					t.Lock()
					notifier.StopFlag = true
					t.Unlock()
				case <-notifier.DeleteDoneChan:
					t.Lock()
					completeCount += 1
					output.Complete = completeCount
					output.Progress = float64(completeCount) / float64(output.FileCount)
					t.Unlock()
					if completeCount == output.FileCount {
						//fmt.Println(" copy complete")
						t.Lock()
						output.Progress = 1
						return
					}
				}
			}
		}()
		err = Delete(src, notifier)
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
