package api

import (
	"github.com/allentom/haruka"
	"net/http"
	"path/filepath"
	"strings"
	"youfile/config"
	"youfile/service"
)

type CreateExtractTaskRequestBody struct {
	Input []struct {
		Input    string `json:"input"`
		Output   string `json:"output"`
		Password string `json:"password"`
		InPlace  bool   `json:"inPlace"`
	} `json:"input"`
}

var newExtractTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody CreateExtractTaskRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	input := make([]*service.ExtractInput, 0)
	realPathMapping := map[string]string{}
	for _, raw := range requestBody.Input {
		if raw.InPlace {
			ext := filepath.Ext(raw.Input)
			dirName := strings.ReplaceAll(filepath.Base(raw.Input), ext, "")
			raw.Output = filepath.Join(filepath.Dir(raw.Input), dirName)
		}
		rawOutput := raw.Output
		if config.Instance.YouPlusPath {
			raw.Input, err = service.GetRealPath(raw.Input, context.Param["token"].(string))
			if err != nil {
				AbortErrorWithStatus(err, context, http.StatusBadRequest)
				return
			}
			raw.Output, err = service.GetRealPath(raw.Input, context.Param["token"].(string))
			if err != nil {
				AbortErrorWithStatus(err, context, http.StatusBadRequest)
				return
			}
		}
		realPathMapping[raw.Output] = rawOutput
		input = append(input, &service.ExtractInput{
			Input:    raw.Input,
			Output:   raw.Output,
			Password: raw.Password,
		})
	}
	task := service.DefaultTask.NewExtractTask(input, service.ExtractTaskOption{
		OnComplete: func(id string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventUnarchiveComplete,
				"id":    id,
			})
		},
		OnFileExtractComplete: func(id string, output string) {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": EventUnarchiveFileComplete,
				"id":    id,
				"path":  realPathMapping[output],
				"dir":   filepath.Dir(realPathMapping[output]),
			})
		},
	}, context.Param["username"].(string))
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
	sourceRealPaths := make([]string, 0)
	for _, source := range requestBody.Sources {
		realPath, err := service.GetRealPath(source, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
		sourceRealPaths = append(sourceRealPaths, realPath)
	}
	rawTarget := requestBody.Target
	requestBody.Target, err = service.GetRealPath(requestBody.Target, context.Param["token"].(string))
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	task := service.DefaultTask.NewArchiveTask(requestBody.Sources, requestBody.Target, func(id string, target string) {
		DefaultNotificationManager.sendJSONToAll(haruka.JSON{
			"event":  EventArchiveComplete,
			"id":     id,
			"target": rawTarget,
		})
	}, context.Param["username"].(string))
	go task.Run()
	context.JSON(task)
}
