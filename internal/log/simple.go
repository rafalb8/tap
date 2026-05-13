package log

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
)

type Simple struct {
	Writer  io.Writer
	Channel chan Frame
}

func (s *Simple) Handle(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	s.Channel <- Frame{
		Timestamp: time.Now(),
		Response:  resp,
	}
	return resp
}

func (s *Simple) Log(ctx context.Context) {
	for {
		select {
		case f := <-s.Channel:
			// [10:31:05] POST → https://api.example.com/v1/auth | 200 OK
			fmt.Fprintf(s.Writer, "[%s] %s\t→ %s | %s\n",
				f.Timestamp.Format(time.TimeOnly),
				f.Request.Method, f.Request.URL, f.Status,
			)
		case <-ctx.Done():
			return
		}
	}
}
