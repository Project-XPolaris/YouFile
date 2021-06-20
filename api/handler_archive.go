package api

import (
	"github.com/allentom/haruka"
	"net/http"
	"path/filepath"
	"strings"
	"youfile/config"
	"youfile/service"
)

type CreateUnarchiveTaskRequestBody struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	InPlace bool   `json:"inPlace"`
}

var newUnarchiveTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateUnarchiveTaskRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	if config.Instance.YouPlusPath {
		requestBody.Source, err = service.GetRealPath(requestBody.Source, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
		if len(requestBody.Target) > 0 && !requestBody.InPlace {
			requestBody.Target, err = service.GetRealPath(requestBody.Target, context.Param["token"].(string))
			if err != nil {
				AbortErrorWithStatus(err, context, http.StatusBadRequest)
				return
			}
		}
	}
	if requestBody.InPlace {
		ext := filepath.Ext(requestBody.Source)
		dirName := strings.ReplaceAll(filepath.Base(requestBody.Source), ext, "")
		requestBody.Target = filepath.Join(filepath.Dir(requestBody.Source), dirName)
	}
	task := service.DefaultTask.NewUnarchiveTask(requestBody.Source, requestBody.Target, func(id string, target string) {
		dir, _ := filepath.Split(target)
		DefaultNotificationManager.sendJSONToAll(haruka.JSON{
			"event":   EventUnarchiveComplete,
			"id":      id,
			"target":  target,
			"context": dir,
		})
	})
	go task.Run()
	context.JSON(task)
}

type CreateArchiveTaskRequestBody struct {
	Sources []string `json:"sources"`
	Target  string   `json:"target"`
}

var newArchiveTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateArchiveTaskRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	task := service.DefaultTask.NewArchiveTask(requestBody.Sources, requestBody.Target, func(id string, target string) {
		DefaultNotificationManager.sendJSONToAll(haruka.JSON{
			"event":  EventArchiveComplete,
			"id":     id,
			"target": target,
		})
	})
	go task.Run()
	context.JSON(task)
}
