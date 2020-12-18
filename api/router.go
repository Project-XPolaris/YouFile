package api

import "github.com/allentom/haruka"

func SetRouter(e *haruka.Engine) {
	e.Router.AddHandler("/path/read", readDirHandler)
	e.Router.AddHandler("/path/copy", copyFileHandler)
	e.Router.AddHandler("/path/remove", deleteFileHandler)
	e.Router.AddHandler("/path/rename", renameFileHandler)
	e.Router.AddHandler("/path/download", downloadFileHandler)
	e.Router.AddHandler("/path/chmod", chmodFileHandler)
	e.Router.AddHandler("/path/search", searchFileHandler)
	e.Router.AddHandler("/path/mkdir", createDirectoryHandler)
	e.Router.AddHandler("/task/search", newSearchFileTaskHandler)
	e.Router.AddHandler("/task/copy", newCopyFileTaskHandler)
	e.Router.AddHandler("/task/delete", newDeleteTaskHandler)
	e.Router.AddHandler("/task/stop", stopTaskHandler)
	e.Router.AddHandler("/task/get", getTaskHandler)
	e.Router.AddHandler("/task/all", getTaskList)
}
