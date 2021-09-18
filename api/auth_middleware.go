package api

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/project-xpolaris/youplustoolkit/youplus/rpc"
	"github.com/sirupsen/logrus"
	"strings"
	"youfile/config"
	"youfile/youplus"
)

type AuthMiddleware struct {
}

func (m *AuthMiddleware) OnRequest(ctx *haruka.Context) {
	rawString := ctx.Request.Header.Get("Authorization")
	ctx.Param["token"] = rawString
	if config.Instance.YouPlusAuth {
		tokenStr := strings.ReplaceAll(rawString, "Bearer ", "")
		if len(tokenStr) == 0 {
			tokenStr = ctx.GetQueryString("token")
		}
		if len(tokenStr) == 0 {
			AbortErrorWithStatus(errors.New("token is empty"), ctx, 403)
			ctx.Interrupt()
			return
		}
		reply, err := youplus.DefaultYouPlusRPCClient.Client.CheckToken(youplus.GenerateRPCTimeoutContext(), &rpc.CheckTokenRequest{Token: &tokenStr})
		if err != nil {
			AbortErrorWithStatus(err, ctx, 403)
			ctx.Interrupt()
			logrus.Error(err)
			return
		}
		if !reply.GetSuccess() {
			AbortErrorWithStatus(errors.New(reply.GetReason()), ctx, 403)
			ctx.Interrupt()
			logrus.Error(reply.GetReason())
			return
		}
		ctx.Param["username"] = reply.GetUsername()
		ctx.Param["uid"] = reply.GetUid()
	} else {
		ctx.Param["username"] = ""
		ctx.Param["uid"] = ""
	}
}
