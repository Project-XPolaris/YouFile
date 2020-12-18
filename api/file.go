package api

import (
	"errors"
	"fmt"
	"github.com/allentom/haruka"
	"net/http"
	"path/filepath"
	"youfile/service"
	"youfile/template"
	"youfile/util"
)

var readDirHandler haruka.RequestHandler = func(context *haruka.Context) {
	readPath := context.GetQueryString("readPath")
	if len(readPath) == 0 {
		readPath = "/"
	}
	items, err := service.ReadDir(util.ConvertPathWithOS(readPath))
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	data := template.NewFileListTemplate(items, readPath)
	err = context.JSON(data)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
}

var copyFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	src := context.GetQueryString("src")
	dest := context.GetQueryString("dest")

	err := service.Copy(util.ConvertPathWithOS(src), util.ConvertPathWithOS(dest), nil)
	if err != nil {
		fmt.Println(err.Error())
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var deleteFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	err := service.DeleteFile(target)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var renameFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	newName := context.GetQueryString("new")
	oldName := context.GetQueryString("old")
	err := service.Rename(util.ConvertPathWithOS(oldName), util.ConvertPathWithOS(newName))
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var downloadFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	targetPath := context.GetQueryString("targetPath")
	context.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(util.ConvertPathWithOS(targetPath))))
	http.ServeFile(context.Writer, context.Request, targetPath)
}

var chmodFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	perm, err := context.GetQueryInt("perm")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.Chmod(util.ConvertPathWithOS(target), perm)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var searchFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	searchPath := context.GetQueryString("searchPath")
	searchKey := context.GetQueryString("searchKey")
	limit, err := context.GetQueryInt("limit")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	items, err := service.SearchFile(util.ConvertPathWithOS(searchPath), searchKey, nil, limit)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	data := template.NewFileListTemplateFromTargetFile(items)
	err = context.JSON(data)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
}

var newSearchFileTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	searchPath := context.GetQueryString("searchPath")
	searchKey := context.GetQueryString("searchKey")
	limit, err := context.GetQueryInt("limit")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	task := service.DefaultTask.NewSearchFileTask(util.ConvertPathWithOS(searchPath), searchKey, limit)
	context.JSON(task)
}

type CreateTaskRequestBody struct {
	List []*service.CopyOption `json:"list"`
}

var newCopyFileTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateTaskRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	for _, option := range requestBody.List {
		option.Src = util.ConvertPathWithOS(option.Src)
		option.Dest = util.ConvertPathWithOS(option.Dest)
	}
	task := service.DefaultTask.NewCopyFileTask(requestBody.List)
	context.JSON(task)
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

var getTaskList haruka.RequestHandler = func(context *haruka.Context) {
	tasks := service.DefaultTask.GetAllTask()
	taskTemplates := make([]*template.TaskTemplate, 0)
	for _, task := range tasks {
		taskTemplates = append(taskTemplates, template.NewTaskTemplate(task))
	}
	context.JSON(map[string]interface{}{
		"result": taskTemplates,
	})
}
var createDirectoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	dirPath := context.GetQueryString("dirPath")
	perm, err := context.GetQueryInt("perm")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.NewDirectory(util.ConvertPathWithOS(dirPath), perm)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

type NewDeleteTaskRequestBody struct {
	List []string `json:"list"`
}

var newDeleteTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody NewDeleteTaskRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	task := service.DefaultTask.NewDeleteFileTask(requestBody.List)
	context.JSON(template.NewTaskTemplate(task))
}
