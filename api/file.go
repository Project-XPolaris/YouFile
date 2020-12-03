package api

import (
	"github.com/allentom/haruka"
	"log"
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
		log.Println(err)
	}
	data := template.NewFileListTemplate(items, readPath)
	err = context.JSON(data)
	if err != nil {
		log.Println(err)
	}
}
