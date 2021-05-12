package api

import (
	"github.com/allentom/haruka"
	"net/http"
	"youfile/service"
)

var clearThumbnailHandler haruka.RequestHandler = func(context *haruka.Context) {
	option := service.ClearThumbnailOption{}
	all := context.GetQueryString("all")
	if all == "1" {
		option.All = true
	}
	err := service.ClearThumbnail(option)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = context.JSON(haruka.JSON{
		"result": "success",
	})
}
