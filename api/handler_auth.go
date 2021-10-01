package api

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/project-xpolaris/youplustoolkit/youplus/rpc"
	"net/http"
	"youfile/util"
	"youfile/youplus"
)

type UserAuthRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var youPlusLoginHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody UserAuthRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	resp, err := youplus.DefaultYouPlusRPCClient.Client.GenerateToken(
		util.GetRPCTimeout(),
		&rpc.GenerateTokenRequest{
			Password: &requestBody.Password,
			Username: &requestBody.Username,
		})
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	if !resp.GetSuccess() {
		AbortErrorWithStatus(errors.New(resp.GetReason()), context, http.StatusBadRequest)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"token":   resp.GetToken(),
		"uid":     resp.GetUid(),
	})
}

var youPlusTokenHandler haruka.RequestHandler = func(context *haruka.Context) {
	token := context.GetQueryString("token")
	resp, err := youplus.DefaultYouPlusRPCClient.Client.CheckToken(
		util.GetRPCTimeout(),
		&rpc.CheckTokenRequest{
			Token: &token,
		},
	)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	if !resp.GetSuccess() {
		AbortErrorWithStatus(errors.New(resp.GetReason()), context, http.StatusBadRequest)
		return
	}
	context.JSON(haruka.JSON{
		"success": resp.Success,
		"uid":     resp.Uid,
	})
}
