package api

import Haruka "github.com/allentom/haruka"

func SetRouter(e *Haruka.Engine) {
	e.Router.AddHandler("/path/read", readDirHandler)
}
