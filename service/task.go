package service

import (
	. "github.com/ahmetb/go-linq/v3"
	"github.com/rs/xid"
	"path/filepath"
	"sync"
	"time"
	"youfile/util"
)

var DefaultTask = NewTaskPool()

const (
	TaskTypeCopy      = "Copy"
	TaskTypeSearch    = "Search"
	TaskTypeDelete    = "Delete"
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
}
type TaskInfo struct {
	Id            string        `json:"id"`
	Type          string        `json:"type"`
	Status        string        `json:"status"`
	Error         error         `json:"error,omitempty"`
	InterruptChan chan struct{} `json:"-"`
	StartTime     *time.Time    `json:"start_time"`
	StopTime      *time.Time    `json:"stop_time"`
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
	Types    []string
	Status   []string
	Order    string
	OrderKey string
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
func (q *TaskQueryBuilder) WithOrder(key string, order string) {
	q.OrderKey = key
	q.Order = order
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

func (t *TaskPool) createTask() TaskInfo {
	task := TaskInfo{
		Id:            xid.New().String(),
		InterruptChan: make(chan struct{}),
	}
	startTime := time.Now()
	task.StartTime = &startTime
	return task
}

type SearchFileOutput struct {
	Files []TargetFile
}
type NewSearchTaskOption struct {
	Src    string
	Key    string
	Limit  int
	OnDone func(id string)
	OnHit  func(id string, path string, name string, itemType string)
}

type SearchFileTask struct {
	TaskInfo
	Output *SearchFileOutput
	sync.Mutex
	Option *NewSearchTaskOption
}

func (t *TaskPool) NewSearchFileTask(option *NewSearchTaskOption) Task {
	taskInfo := t.createTask()
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

type NewCopyTaskOption struct {
	Options []*CopyOption
	OnDone  func(id string)
}
type CopyOption struct {
	Src        string          `json:"src"`
	Dest       string          `json:"dest"`
	OnComplete func(id string) `json:"-"`
}
type CopyTask struct {
	TaskInfo
	Output *CopyFileTaskOutput
	sync.Mutex
	Option *NewCopyTaskOption
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

func (t *TaskPool) NewCopyTask(option *NewCopyTaskOption) Task {
	taskInfo := t.createTask()
	taskInfo.Type = TaskTypeCopy
	taskInfo.Status = TaskStateRunning
	task := CopyTask{
		TaskInfo: taskInfo,
		Output: &CopyFileTaskOutput{
			List: option.Options,
		},
		Option: option,
	}
	t.Lock()
	t.Tasks = append(t.Tasks, &task)
	t.Unlock()
	return &task
}
func (t *CopyTask) Run() {
	// analyze
	infos := make([]*CopyAnalyzeResult, 0)
	for _, option := range t.Option.Options {
		copyInfo, err := analyzeSource(option.Src)
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
	t.Output.TotalLength = 0
	for _, info := range infos {
		t.Output.FileCount += info.FileCount
		t.Output.TotalLength += info.TotalSize
	}
	t.Status = TaskStateRunning
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
		var completeCount = 0
		ticker := time.NewTicker(2 * time.Second)
		var lastComplete int64 = 0
		for {
			select {
			case currentFile := <-notifier.CurrentFileChan:
				t.Lock()
				t.Output.CurrentCopy = currentFile
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
		err := Copy(option.Src, option.Dest, notifier)
		if err == util.CopyInterrupt {
			break
		}
		if err != nil {
			t.Lock()
			t.Error = err
			t.UpdateStopTime()
			t.Unlock()
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
		t.Option.OnDone(t.Id)
	}
}

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
}

func (t *TaskPool) NewDeleteFileTask(option *NewDeleteFileTaskOption) Task {
	taskInfo := t.createTask()
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
