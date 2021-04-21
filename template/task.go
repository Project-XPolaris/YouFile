package template

import (
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
}

func NewTaskTemplate(task service.Task) *TaskTemplate {
	formatString := "2006-01-02 15:04:05"
	template := &TaskTemplate{
		Id:     task.GetId(),
		Type:   task.GetType(),
		Status: task.GetStatus(),
		Output: SerializeTaskOutput(task),
		Error:  task.GetError(),
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
		return SerializeSearchFileOutput(v.Output)
	case *service.CopyTask:
		return v.Output
	case *service.DeleteFileTask:
		return v.Output
	case *service.ArchiveTask:
		return map[string]interface{}{}
	case *service.UnarchiveTask:
		return map[string]interface{}{}
	default:
		return data
	}
}
func SerializeSearchFileOutput(data *service.SearchFileOutput) interface{} {
	return NewFileListTemplateFromTargetFile(data.Files)
}
