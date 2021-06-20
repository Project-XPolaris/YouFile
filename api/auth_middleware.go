package api

import (
	"github.com/allentom/haruka"
)

type AuthMiddleware struct {
}

func (m *AuthMiddleware) OnRequest(ctx *haruka.Context) {
	rawString := ctx.Request.Header.Get("Authorization")
	ctx.Param["token"] = rawString
}
