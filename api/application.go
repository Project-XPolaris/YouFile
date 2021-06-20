package api

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
)

func RunApiService() {
	engine := haruka.NewEngine()
	engine.UseMiddleware(middleware.NewLoggerMiddleware())
	engine.UseMiddleware(&AuthMiddleware{})
	SetRouter(engine)
	engine.RunAndListen(":8300")
}
