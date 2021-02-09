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

var newSearchFileTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	searchPath := context.GetQueryString("searchPath")
	searchKey := context.GetQueryString("searchKey")
	limit, err := context.GetQueryInt("limit")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	task := service.DefaultTask.NewSearchFileTask(&service.NewSearchTaskOption{
		Src:   searchPath,
		Key:   searchKey,
		Limit: limit,
		OnDone: func(id string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": "SearchTaskComplete",
				"id":    id,
			})
		},
		OnHit: func(id string, path string, name string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": "SearchHit",
				"id":    id,
				"path":  path,
				"name":  name,
			})
		},
	})
	task.Run()
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
		option.OnComplete = func(id string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": "CopyItemComplete",
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
				"event": "CopyTaskComplete",
				"id":    id,
			})
		},
	})
	task.Run()
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
	queryBuilder := service.TaskQueryBuilder{}
	queryBuilder.WithOrder(context.GetQueryString("orderKey"), context.GetQueryString("order"))
	queryBuilder.WithStatus(context.GetQueryStrings("status")...)
	queryBuilder.WithTypes(context.GetQueryStrings("type")...)
	tasks := queryBuilder.Query()
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
	task := service.DefaultTask.NewDeleteFileTask(&service.NewDeleteFileTaskOption{
		Src: requestBody.List,
		OnDone: func(id string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": "DeleteTaskDone",
				"id":    id,
			})
		},
		OnItemComplete: func(id string, src string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": "DeleteItemComplete",
				"id":    id,
				"src":   src,
			})
		},
	})
	task.Run()
	context.JSON(template.NewTaskTemplate(task))
}

var mountCifsHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody service.MountCIFSOption
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.MountCIFS(requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var umountHandler haruka.RequestHandler = func(context *haruka.Context) {
	dirPath := context.GetQueryString("dirPath")
	err := service.UmountFS(dirPath)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var fstabMountListHandler haruka.RequestHandler = func(context *haruka.Context) {
	data := service.DefaultFstab.Mounts
	context.JSON(template.MountTemplateFromList(data))
}

var fstabAddMountHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody service.AddMountOption
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	service.DefaultFstab.AddMount(&requestBody)
	err = service.DefaultFstab.Save()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = service.DefaultFstab.Reload()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var fstabRemoveMountHandler haruka.RequestHandler = func(context *haruka.Context) {
	dirPath := context.GetQueryString("dirPath")
	err := service.DefaultFstab.RemoveMount(dirPath)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = service.DefaultFstab.Save()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = service.DefaultFstab.Reload()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var fstabReMountHandler haruka.RequestHandler = func(context *haruka.Context) {
	err := service.DefaultFstab.Save()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = service.DefaultFstab.Reload()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

type OSInfoResponse struct {
	RootPaths []service.RootPath `json:"root_paths,omitempty"`
	Sep       string             `json:"sep" json:"sep,omitempty"`
}

var readOSInfoDirHandler haruka.RequestHandler = func(context *haruka.Context) {
	paths, err := service.GetStartPath()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	info := OSInfoResponse{
		RootPaths: paths,
		Sep:       string(filepath.Separator),
	}
	context.JSON(info)
}

var getFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	targetPath := context.GetQueryString("target")
	http.ServeFile(context.Writer, context.Request, targetPath)
}
