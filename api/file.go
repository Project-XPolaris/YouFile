package api

import (
	"errors"
	"fmt"
	"github.com/allentom/haruka"
	"net/http"
	"path/filepath"
	"youfile/service"
	"youfile/template"
)

var readDirHandler haruka.RequestHandler = func(context *haruka.Context) {
	readPath := context.GetQueryString("readPath")
	if len(readPath) == 0 {
		readPath = "./"
	}
	items, err := service.ReadDir(readPath)
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

	err := service.Copy(src, dest, nil)
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
	err := service.Rename(oldName, newName)
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
	context.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(targetPath)))
	http.ServeFile(context.Writer, context.Request, targetPath)
}

var chmodFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	perm, err := context.GetQueryInt("perm")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.Chmod(target, perm)
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
	items, err := service.SearchFile(searchPath, searchKey)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	data := template.NewFileListTemplate(items, searchPath)
	err = context.JSON(data)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
}

var newSearchFileTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	searchPath := context.GetQueryString("searchPath")
	searchKey := context.GetQueryString("searchKey")
	task := service.DefaultTask.NewScanFileTask(searchPath, searchKey)
	context.JSON(task)
}

var newCopyFileTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	src := context.GetQueryString("src")
	dest := context.GetQueryString("dest")
	task := service.DefaultTask.NewCopyFileTask(src, dest)
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

var createDirectoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	dirPath := context.GetQueryString("dirPath")
	perm, err := context.GetQueryInt("perm")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.NewDirectory(dirPath, perm)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}
