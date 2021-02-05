package api

import (
	"github.com/allentom/haruka"
	"log"
)

func AbortErrorWithStatus(err error, context *haruka.Context, status int) {
	log.Println(err)
	context.Writer.Header().Set("Content-type", "application/json")
	context.Writer.WriteHeader(status)
	context.JSON(map[string]interface{}{
		"success": false,
		"reason":  err.Error(),
	})
}
