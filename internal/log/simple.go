package log

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
)

type Simple struct {
	Writer io.Writer
}

func (s *Simple) Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	panic("TODO")
}

func (s *Simple) Response(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if resp == nil {
		panic("Response is nil")
	}
	// [10:31:05] POST → https://api.example.com/v1/auth | 200 OK
	fmt.Fprintf(s.Writer, "[%s] %s\t→ %s | %s\n",
		time.Now().Format(time.TimeOnly),
		resp.Request.Method, resp.Request.URL, resp.Status,
	)

	return resp
}
