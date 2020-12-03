package api

import "github.com/allentom/haruka"

func RunApiService() {
	engine := haruka.NewEngine()
	SetRouter(engine)
	engine.RunAndListen(":8100")
}
