package api

import (
	"github.com/allentom/haruka"
	"youfile/config"
)

var infoHandler haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(haruka.JSON{
		"name":        "YouFile Service",
		"auth":        config.Instance.YouPlusPath,
		"youPlusPath": config.Instance.YouPlusPath,
		"success":     true,
	})
}
