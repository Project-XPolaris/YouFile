package api

import (
	gocontext "context"
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/allentom/haruka"
	"github.com/d-tux/go-fstab"
	"github.com/project-xpolaris/youplustoolkit/youplus/rpc"
	"net/http"
	"path/filepath"
	"youfile/config"
	"youfile/service"
	"youfile/template"
	"youfile/util"
	"youfile/youplus"
)

var (
	FeatureNotEnableError = errors.New("feature not enable")
)
var readDirHandler haruka.RequestHandler = func(context *haruka.Context) {
	readPath := context.GetQueryString("readPath")
	realPath, err := service.GetRealPath(readPath, context.Param["token"].(string))
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	thumbnail := context.GetQueryString("thumbnail")

	items, err := service.ReadDir(util.ConvertPathWithOS(realPath))
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	if thumbnail != "0" && config.Instance.Thumbnails {
		go service.GenerateImageThumbnail(realPath, func() {
			DefaultNotificationManager.sendJSONToAll(haruka.JSON{
				"event": GenerateThumbnailComplete,
				"path":  readPath,
			})
		})
	}

	data := template.NewFileListTemplate(items, readPath, realPath)
	err = context.JSON(data)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
}
var getFolderDataset haruka.RequestHandler = func(context *haruka.Context) {
	if !config.Instance.YouPlusZFS {
		AbortErrorWithStatus(FeatureNotEnableError, context, http.StatusForbidden)
		return
	}
	path := context.GetQueryString("path")
	reply, err := youplus.DefaultYouPlusRPCClient.Client.GetDatasetInfo(
		gocontext.Background(),
		&rpc.GetDatasetInfoRequest{Dataset: &path},
	)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	snapshots := make([]template.DatasetSnapshot, 0)
	for _, snapshot := range reply.Snapshots {
		snapshots = append(snapshots, template.DatasetSnapshot{Name: snapshot.GetName()})
	}
	context.JSON(haruka.JSON{
		"success":   true,
		"snapshots": snapshots,
	})
}
var createFolderDataset haruka.RequestHandler = func(context *haruka.Context) {
	if !config.Instance.YouPlusZFS {
		AbortErrorWithStatus(FeatureNotEnableError, context, http.StatusForbidden)
		return
	}
	path := context.GetQueryString("path")
	reply, err := youplus.DefaultYouPlusRPCClient.Client.CreateDataset(
		gocontext.Background(),
		&rpc.CreateDatasetRequest{Path: &path},
	)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": reply.Success,
	})
}
var deleteFolderDataset haruka.RequestHandler = func(context *haruka.Context) {
	if !config.Instance.YouPlusZFS {
		AbortErrorWithStatus(FeatureNotEnableError, context, http.StatusForbidden)
		return
	}
	path := context.GetQueryString("path")
	reply, err := youplus.DefaultYouPlusRPCClient.Client.DeleteDataset(
		gocontext.Background(),
		&rpc.DeleteDatasetRequest{Path: &path},
	)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": reply.Success,
	})
}
var createSnapshot haruka.RequestHandler = func(context *haruka.Context) {
	if !config.Instance.YouPlusZFS {
		AbortErrorWithStatus(FeatureNotEnableError, context, http.StatusForbidden)
		return
	}
	path := context.GetQueryString("path")
	name := context.GetQueryString("name")
	reply, err := youplus.DefaultYouPlusRPCClient.Client.CreateSnapshot(
		gocontext.Background(),
		&rpc.CreateSnapshotRequest{Dataset: &path, Snapshot: &name},
	)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": reply.Success,
	})
}
var deleteSnapshot haruka.RequestHandler = func(context *haruka.Context) {
	if !config.Instance.YouPlusZFS {
		AbortErrorWithStatus(FeatureNotEnableError, context, http.StatusForbidden)
		return
	}
	path := context.GetQueryString("path")
	name := context.GetQueryString("name")
	reply, err := youplus.DefaultYouPlusRPCClient.Client.DeleteSnapshot(
		gocontext.Background(),
		&rpc.DeleteSnapshotRequest{Dataset: &path, Snapshot: &name},
	)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": reply.Success,
	})
}
var rollbackDataset haruka.RequestHandler = func(context *haruka.Context) {
	if !config.Instance.YouPlusZFS {
		AbortErrorWithStatus(FeatureNotEnableError, context, http.StatusForbidden)
		return
	}
	path := context.GetQueryString("path")
	name := context.GetQueryString("name")
	reply, err := youplus.DefaultYouPlusRPCClient.Client.RollbackDataset(
		gocontext.Background(),
		&rpc.RollbackDatasetRequest{Dataset: &path, Snapshot: &name},
	)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": reply.Success,
	})
}
var copyFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	src := context.GetQueryString("src")
	dest := context.GetQueryString("dest")
	if config.Instance.YouPlusPath {
		src, err = service.GetRealPath(src, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
		dest, err = service.GetRealPath(dest, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
	}
	err = service.Copy(util.ConvertPathWithOS(src), util.ConvertPathWithOS(dest), nil, "rename")
	if err != nil {
		fmt.Println(err.Error())
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var deleteFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	target := context.GetQueryString("target")
	if config.Instance.YouPlusPath {
		target, err = service.GetRealPath(target, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
	}
	err = service.DeleteFile(target)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var renameFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	newName := context.GetQueryString("new")
	oldName := context.GetQueryString("old")
	if config.Instance.YouPlusPath {
		newName, err = service.GetRealPath(newName, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
		oldName, err = service.GetRealPath(oldName, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
	}
	err = service.Rename(util.ConvertPathWithOS(oldName), util.ConvertPathWithOS(newName))
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var downloadFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	targetPath := context.GetQueryString("targetPath")
	if config.Instance.YouPlusPath {
		targetPath, err = service.GetRealPath(targetPath, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
	}
	context.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(util.ConvertPathWithOS(targetPath))))
	http.ServeFile(context.Writer, context.Request, targetPath)
}

var chmodFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	target := context.GetQueryString("target")
	perm, err := context.GetQueryInt("perm")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	if config.Instance.YouPlusPath {
		target, err = service.GetRealPath(target, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
	}
	err = service.Chmod(util.ConvertPathWithOS(target), perm)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

var createDirectoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	dirPath := context.GetQueryString("dirPath")
	perm, err := context.GetQueryInt("perm")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	realPath, err := service.GetRealPath(dirPath, context.Param["token"].(string))
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.NewDirectory(realPath, perm)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = context.JSON(map[string]interface{}{
		"result": "success",
	})
}

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

type OSInfoResponse struct {
	RootPaths []service.RootPath `json:"root_paths,omitempty"`
	Sep       string             `json:"sep" json:"sep,omitempty"`
}

var readOSInfoDirHandler haruka.RequestHandler = func(context *haruka.Context) {
	paths, err := service.GetStartPath(context.Param["token"].(string))
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	info := OSInfoResponse{
		RootPaths: paths,
		Sep:       string(filepath.Separator),
	}
	context.JSON(info)
}

var getFileHandler haruka.RequestHandler = func(context *haruka.Context) {
	var err error
	targetPath := context.GetQueryString("target")
	if config.Instance.YouPlusPath {
		targetPath, err = service.GetRealPath(targetPath, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
	}
	http.ServeFile(context.Writer, context.Request, targetPath)
}
