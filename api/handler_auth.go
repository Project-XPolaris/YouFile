package api

import (
	"github.com/allentom/haruka"
	"net/http"
	"net/http/httputil"
	"net/url"
	"youfile/config"
)

var youPlusLoginHandler haruka.RequestHandler = func(context *haruka.Context) {
	url, err := url.Parse(config.Instance.YouPlusUrl)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	request := context.Request
	request.URL.Path = "/user/auth"
	proxy.ServeHTTP(context.Writer, request)
}
