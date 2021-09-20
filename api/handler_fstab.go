package api

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/haruka"
	"github.com/d-tux/go-fstab"
	"net/http"
	"youfile/config"
	"youfile/service"
	"youfile/template"
)

var fstabMountListHandler haruka.RequestHandler = func(context *haruka.Context) {
	data := service.DefaultFstab.Mounts
	linq.From(data).Where(func(i interface{}) bool {
		for _, point := range config.Instance.MountPoints {
			if point == i.(*fstab.Mount).File {
				return true
			}
		}
		return false
	}).ToSlice(&data)
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
	config.Instance.MountPoints = append(config.Instance.MountPoints, requestBody.File)
	err = config.SaveMounts()
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
	linq.From(config.Instance.MountPoints).Where(func(i interface{}) bool {
		return i.(string) != dirPath
	}).ToSlice(&config.Instance.MountPoints)
	err = config.SaveMounts()
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
