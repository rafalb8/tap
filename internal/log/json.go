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
	if req == nil {
		fmt.Fprintln(j.Writer, "Request is nil")
		return nil, nil
	}

	ts := time.Now()
	resp, err := ctx.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(j.Writer, "Response error: %v\n", err)
		return req, nil
	}
	latency := time.Since(ts)

	fmt.Fprintf(j.Writer,
		`{"timestamp":"%s", "method": "%s", "url":"%s", "status": "%s", "latency": "%s"}`+"\n",
		time.Now().Format(time.RFC3339),
		resp.Request.Method, resp.Request.URL, resp.Status, latency,
	)

	return req, resp
}
