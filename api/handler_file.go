package api

import (
	"github.com/allentom/haruka"
	"net/http"
	"path/filepath"
	"youfile/service"
)

type NewTextFileRequestBody struct {
	FilePath string `json:"filePath"`
}

var newTextFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody NewTextFileRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.NewTextFile(requestBody.FilePath)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = context.JSON(haruka.JSON{
		"result": "success",
	})
}

type WriteFileRequestBody struct {
	FilePath string `json:"filePath"`
	Content  string `json:"content"`
}

var writeFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody WriteFileRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.WriteTextFile(requestBody.FilePath, requestBody.Content)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = context.JSON(haruka.JSON{
		"result": "success",
	})
}

var readAsTextFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("path")
	text, err := service.ReadFileAsString(path)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = context.JSON(haruka.JSON{
		"result":  "success",
		"content": text,
	})
}

var getFileThumbnailHandler haruka.RequestHandler = func(context *haruka.Context) {
	path := context.GetQueryString("name")
	http.ServeFile(context.Writer, context.Request, filepath.Join("./thumbnails", path))
}
