package api

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
	"github.com/project-xpolaris/youplustoolkit/youlink"
	"log"
	"youfile/config"
)

func RunApiService() {
	engine := haruka.NewEngine()
	engine.UseMiddleware(middleware.NewLoggerMiddleware())
	engine.UseMiddleware(&AuthMiddleware{})
	SetRouter(engine)
	if config.Instance.YouLink.Enable {
		service := youlink.NewService(config.Instance.YouLink.Url, config.Instance.YouLink.ServiceUrl)
		service.AddFunction(
			&youLinkMoveFileFunction,
		)
		service.RegisterHarukaHandler(engine)
		err := service.RegisterFunction()
		if err != nil {
			log.Fatal(err)
		}
	}
	engine.RunAndListen(config.Instance.Addr)
}
