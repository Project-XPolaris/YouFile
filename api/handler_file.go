package api

import (
	gocontext "context"
	"errors"
	"fmt"
	"github.com/allentom/haruka"
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
	if config.Instance.YouPlusPath {
		requestBody.FilePath, err = service.GetRealPath(requestBody.FilePath, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
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
	if config.Instance.YouPlusPath {
		requestBody.FilePath, err = service.GetRealPath(requestBody.FilePath, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
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
	var err error
	path := context.GetQueryString("path")
	if config.Instance.YouPlusPath {
		path, err = service.GetRealPath(path, context.Param["token"].(string))
		if err != nil {
			AbortErrorWithStatus(err, context, http.StatusBadRequest)
			return
		}
	}
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
