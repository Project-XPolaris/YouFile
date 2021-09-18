package api

import (
	"errors"
	"github.com/allentom/haruka"
	"net/http"
	"youfile/config"
	"youfile/service"
	"youfile/template"
	"youfile/util"
)

var getTaskList haruka.RequestHandler = func(context *haruka.Context) {
	queryBuilder := service.TaskQueryBuilder{}
	queryBuilder.
		WithOrder(context.GetQueryString("orderKey"), context.GetQueryString("order")).
		WithStatus(context.GetQueryStrings("status")...).
		WithTypes(context.GetQueryStrings("type")...).
		InUser(context.Param["username"].(string))
	tasks := queryBuilder.Query()
	taskTemplates := make([]*template.TaskTemplate, 0)
	for _, task := range tasks {
		taskTemplates = append(taskTemplates, template.NewTaskTemplate(task))
	}
	context.JSON(map[string]interface{}{
		"result": taskTemplates,
	})
}
var getTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	taskId := context.GetQueryString("taskId")
	task := service.DefaultTask.GetTask(taskId)
	if task == nil {
		AbortErrorWithStatus(errors.New("task not found"), context, http.StatusNotFound)
		return
	}
	context.JSON(template.NewTaskTemplate(task))
}

var stopTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	taskId := context.GetQueryString("taskId")
	service.DefaultTask.StopTask(taskId)
	context.JSON(map[string]interface{}{
		"result": "success",
	})
}

type NewDeleteTaskRequestBody struct {
	List []string `json:"list"`
}

var newDeleteTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	username := context.Param["username"].(string)
	var requestBody NewDeleteTaskRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	realDeletePath := make([]string, 0)
	realPathMapping := map[string]string{}
	for _, deletePath := range requestBody.List {
		realPath, err := service.GetRealPath(deletePath, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
		realDeletePath = append(realDeletePath, realPath)
		realPathMapping[realPath] = deletePath
	}
	task := service.DefaultTask.NewDeleteFileTask(&service.NewDeleteFileTaskOption{
		Src: realDeletePath,
		OnDone: func(id string) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventDeleteTaskDone,
				"id":    id,
			}, username)
		},
		OnItemComplete: func(id string, src string) {
			DefaultNotificationManager.sendJSONToUser(haruka.JSON{
				"event": EventDeleteItemComplete,
				"id":    id,
				"src":   realPathMapping[src],
			}, username)
		},
		Username: username,
	})
	go task.Run()
	context.JSON(template.NewTaskTemplate(task))
}

type CreateCopyTaskRequestBody struct {
	List      []*service.CopyOption `json:"list"`
	Duplicate string                `json:"duplicate"`
}

var newCopyFileTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateCopyTaskRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	for _, option := range requestBody.List {
		realSrc, err := service.GetRealPath(option.Src, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
		realDest, err := service.GetRealPath(option.Dest, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
		option.Src = util.ConvertPathWithOS(realSrc)
		option.Dest = util.ConvertPathWithOS(realDest)
		option.OnComplete = func(id string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventCopyItemComplete,
				"id":    id,
				"src":   option.Src,
				"dest":  option.Dest,
			})
		}
	}
	task := service.DefaultTask.NewCopyTask(&service.NewCopyTaskOption{
		Options: requestBody.List,
		OnDone: func(id string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventCopyTaskComplete,
				"id":    id,
			})
		},
		Username:    context.Param["username"].(string),
		OnDuplicate: requestBody.Duplicate,
	})
	go task.Run()
	context.JSON(template.NewTaskTemplate(task))
}

var newSearchFileTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	searchPath := context.GetQueryString("searchPath")
	searchKey := context.GetQueryString("searchKey")
	limit, err := context.GetQueryInt("limit")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	realPath := searchPath
	if config.Instance.YouPlusPath {
		realPath, err = service.GetRealPath(searchPath, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
	}
	task := service.DefaultTask.NewSearchFileTask(&service.NewSearchTaskOption{
		Src:   realPath,
		Key:   searchKey,
		Limit: limit,
		OnDone: func(id string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventSearchTaskComplete,
				"id":    id,
			})
		},
		PathTrans: searchPath,
	})
	go task.Run()
	taskTemplate := template.NewTaskTemplate(task)
	context.JSON(taskTemplate)
}
