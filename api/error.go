package api

import "github.com/allentom/haruka"

func AbortErrorWithStatus(err error, context *haruka.Context, status int) {
	context.Writer.WriteHeader(status)
	context.JSON(map[string]interface{}{
		"success": false,
		"reason":  err.Error(),
	})
}
