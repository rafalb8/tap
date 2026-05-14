package log

import (
	"net/http"

	"github.com/elazarl/goproxy"
)

type Logger interface {
	Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response)
	Response(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response
}
