package template

import (
	"youfile/service"
)

type TaskTemplate struct {
	Id     string      `json:"id"`
	Type   string      `json:"type"`
	Status string      `json:"status"`
	Output interface{} `json:"output,omitempty"`
	Error  error       `json:"error,omitempty"`
}

func NewTaskTemplate(task *service.Task) *TaskTemplate {
	return &TaskTemplate{
		Id:     task.Id,
		Type:   task.Type,
		Status: task.Status,
		Output: SerializeTaskOutput(task.Output),
		Error:  task.Error,
	}
}
func SerializeTaskOutput(data interface{}) interface{} {
	switch v := data.(type) {
	case *service.SearchFileOutput:
		return SerializeSearchFileOutput(v)
	default:
		return data
	}
}
func SerializeSearchFileOutput(data *service.SearchFileOutput) interface{} {
	return NewFileListTemplateFromTargetFile(data.Files)
}
