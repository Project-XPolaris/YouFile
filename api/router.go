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
	e.Router.AddHandler("/task/search", newSearchFileTaskHandler)
	e.Router.AddHandler("/task/copy", newCopyFileTaskHandler)
	e.Router.AddHandler("/task/get", getTaskHandler)
}
