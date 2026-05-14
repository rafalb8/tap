package log

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
)

type Json struct {
	Writer io.Writer
}

func (j *Json) Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	panic("TODO")
}

func (j *Json) Response(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if resp == nil {
		panic("Response is nil")
	}

	fmt.Fprintf(j.Writer,
		`{"timestamp":"%s", "method": "%s", "url":"%s", "status": "%s"}`+"\n",
		time.Now().Format(time.RFC3339),
		resp.Request.Method, resp.Request.URL, resp.Status,
	)
	return resp
}
