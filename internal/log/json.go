package log

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
)

type Json struct {
	Writer  io.Writer
	Channel chan Frame
}

func (s *Json) Handle(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if resp == nil {
		panic("Response is nil")
	}

	s.Channel <- Frame{
		Timestamp: time.Now(),
		Response:  resp,
	}
	return resp
}

func (s *Json) Log(ctx context.Context) {
	for {
		select {
		case f := <-s.Channel:
			// [10:31:05] POST → https://api.example.com/v1/auth | 200 OK (45ms)
			fmt.Fprintf(s.Writer, `{"timestamp":"%s", "method": "%s", "url":"%s", "status": "%s"}`,
				f.Timestamp.Format(time.RFC3339),
				f.Request.Method, f.Request.URL, f.Status,
			)
			fmt.Fprintln(s.Writer)
		case <-ctx.Done():
			return
		}
	}
}
