package template

import (
	"path/filepath"
	"youfile/service"
)

type TaskTemplate struct {
	Id        string      `json:"id"`
	Type      string      `json:"type"`
	Status    string      `json:"status"`
	Output    interface{} `json:"output,omitempty"`
	Error     error       `json:"error,omitempty"`
	StartTime string      `json:"start_time,omitempty"`
	StopTime  string      `json:"stop_time,omitempty"`
	Username  string      `json:"username"`
}

func NewTaskTemplate(task service.Task) *TaskTemplate {
	formatString := "2006-01-02 15:04:05"
	template := &TaskTemplate{
		Id:       task.GetId(),
		Type:     task.GetType(),
		Status:   task.GetStatus(),
		Output:   SerializeTaskOutput(task),
		Error:    task.GetError(),
		Username: task.GetUsername(),
	}
	if task.GetStartTime() != nil {
		template.StartTime = task.GetStartTime().Format(formatString)
	}
	if task.GetStopTime() != nil {
		template.StopTime = task.GetStopTime().Format(formatString)
	}
	return template
}
func SerializeTaskOutput(data interface{}) interface{} {
	switch v := data.(type) {
	case *service.SearchFileTask:
		return SerializeSearchFileOutput(v.Output, v.Option.Src)
	case *service.CopyTask:
		return SerializeCopyFileOutput(v)
	case *service.DeleteFileTask:
		return SerializeDeleteFileOutput(v)
	case *service.ArchiveTask:
		return map[string]interface{}{}
	case *service.ExtractTask:
		return SerializeExtractOutput(v)
	case *service.MoveTask:
		return SerializeMoveFileOutput(v)
	default:
		return data
	}
}
func SerializeSearchFileOutput(data *service.SearchFileOutput, src string) interface{} {
	return NewFileListTemplateFromTargetFile(data.Files, src)
}

func SerializeDeleteFileOutput(data *service.DeleteFileTask) interface{} {
	template := DeleteFileOutputTemplate{}
	template.Serialize(data)
	return template
}

type DeleteFile struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Directory string `json:"directory"`
}
type DeleteFileOutputTemplate struct {
	FileCount     int          `json:"fileCount"`
	Complete      int          `json:"complete"`
	Progress      float64      `json:"progress"`
	Speed         int          `json:"speed"`
	CurrentDelete string       `json:"currentDelete"`
	Files         []DeleteFile `json:"files"`
}

func (t *DeleteFileOutputTemplate) Serialize(task service.Task) {
	deleteTask := task.(*service.DeleteFileTask)
	t.Complete = deleteTask.Output.Complete
	t.FileCount = deleteTask.Output.FileCount
	t.CurrentDelete = deleteTask.Output.CurrentDelete
	t.Progress = deleteTask.Output.Progress
	t.Speed = deleteTask.Output.Speed
	t.Files = []DeleteFile{}
	for _, deleteSource := range deleteTask.Output.Src {
		displayPath := deleteTask.Option.DisplaySrcMapping[deleteSource]
		t.Files = append(t.Files, DeleteFile{
			Path:      displayPath,
			Name:      filepath.Base(displayPath),
			Directory: filepath.Dir(displayPath),
		})
	}
}

func SerializeCopyFileOutput(data *service.CopyTask) interface{} {
	template := CopyFileOutputTemplate{}
	template.Serialize(data)
	return template
}

type CopyFile struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Directory string `json:"directory"`
}
type CopyOption struct {
	Source CopyFile `json:"source"`
	Dest   CopyFile `json:"dest"`
}
type CopyFileOutputTemplate struct {
	TotalLength    int64        `json:"totalLength"`
	FileCount      int          `json:"fileCount"`
	Complete       int          `json:"complete"`
	CompleteLength int64        `json:"completeLength"`
	Files          []CopyOption `json:"files"`
	CurrentCopy    string       `json:"currentCopy"`
	Progress       float64      `json:"progress"`
	Speed          int64        `json:"speed"`
}

func (t *CopyFileOutputTemplate) Serialize(task service.Task) {
	copyTask := task.(*service.CopyTask)
	t.TotalLength = copyTask.Output.TotalLength
	t.CompleteLength = copyTask.Output.CompleteLength
	t.Complete = copyTask.Output.Complete
	t.FileCount = copyTask.Output.FileCount
	t.CurrentCopy = copyTask.Output.CurrentCopy
	t.Progress = copyTask.Output.Progress
	t.Speed = copyTask.Output.Speed
	t.Files = []CopyOption{}
	for _, copySource := range copyTask.Output.List {
		displaySourcePath := copyTask.Option.DisplayPath[copySource.Src]
		displayDestPath := copyTask.Option.DisplayPath[copySource.Dest]
		t.Files = append(t.Files, CopyOption{
			Source: CopyFile{
				Path:      displaySourcePath,
				Name:      filepath.Base(displaySourcePath),
				Directory: filepath.Dir(displaySourcePath),
			},
			Dest: CopyFile{
				Path:      displayDestPath,
				Name:      filepath.Base(displayDestPath),
				Directory: filepath.Dir(displayDestPath),
			},
		})
	}
}

func SerializeExtractOutput(data *service.ExtractTask) interface{} {
	template := ExtractOutputTemplate{}
	template.Serialize(data)
	return template
}

type ExtractFile struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Directory string `json:"directory"`
}
type ExtractOption struct {
	Source        CopyFile `json:"source"`
	Dest          string   `json:"dest"`
	DestDirectory string   `json:"destDirectory"`
}
type ExtractOutputTemplate struct {
	Complete int             `json:"complete"`
	Total    int             `json:"total"`
	Files    []ExtractOption `json:"files"`
}

func (t *ExtractOutputTemplate) Serialize(task service.Task) {
	extractTask := task.(*service.ExtractTask)
	t.Complete = extractTask.Output.Complete
	t.Total = extractTask.Output.Total
	t.Files = []ExtractOption{}
	for _, extractOption := range extractTask.Input {
		displaySourcePath := extractTask.Option.DisplayPath[extractOption.Input]
		displayDestPath := extractTask.Option.DisplayPath[extractOption.Output]
		t.Files = append(t.Files, ExtractOption{
			Source: CopyFile{
				Path:      displaySourcePath,
				Name:      filepath.Base(displaySourcePath),
				Directory: filepath.Dir(displaySourcePath),
			},
			Dest:          displayDestPath,
			DestDirectory: filepath.Dir(displayDestPath),
		})
	}
}

func SerializeMoveFileOutput(data *service.MoveTask) interface{} {
	template := MoveFileOutputTemplate{}
	template.Serialize(data)
	return template
}

type MoveFile struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Directory string `json:"directory"`
}
type MoveOption struct {
	Source MoveFile `json:"source"`
	Dest   MoveFile `json:"dest"`
}
type MoveFileOutputTemplate struct {
	TotalLength    int64        `json:"totalLength"`
	FileCount      int          `json:"fileCount"`
	Complete       int          `json:"complete"`
	CompleteLength int64        `json:"completeLength"`
	Files          []MoveOption `json:"files"`
	CurrentMove    string       `json:"currentMove"`
	Progress       float64      `json:"progress"`
	Speed          int64        `json:"speed"`
}

func (t *MoveFileOutputTemplate) Serialize(task service.Task) {
	moveTask := task.(*service.MoveTask)
	t.TotalLength = moveTask.Output.TotalLength
	t.CompleteLength = moveTask.Output.CompleteLength
	t.Complete = moveTask.Output.Complete
	t.FileCount = moveTask.Output.FileCount
	t.CurrentMove = moveTask.Output.CurrentMove
	t.Progress = moveTask.Output.Progress
	t.Speed = moveTask.Output.Speed
	t.Files = []MoveOption{}
	for _, copySource := range moveTask.Output.List {
		displaySourcePath := moveTask.Option.DisplayPath[copySource.Src]
		displayDestPath := moveTask.Option.DisplayPath[copySource.Dest]
		t.Files = append(t.Files, MoveOption{
			Source: MoveFile{
				Path:      displaySourcePath,
				Name:      filepath.Base(displaySourcePath),
				Directory: filepath.Dir(displaySourcePath),
			},
			Dest: MoveFile{
				Path:      displayDestPath,
				Name:      filepath.Base(displayDestPath),
				Directory: filepath.Dir(displayDestPath),
			},
		})
	}
}
