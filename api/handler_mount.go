package api

import (
	"github.com/allentom/haruka"
	"net/http"
	"youfile/config"
	"youfile/service"
)

var mountCifsHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody service.MountCIFSOption
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	if config.Instance.YouPlusPath {
		requestBody.MountPath, err = service.GetRealPath(requestBody.MountPath, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
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
