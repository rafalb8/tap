package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
	"unsafe"

	"github.com/elazarl/goproxy"
)

type Standard struct {
	Writer  io.Writer
	Channel chan Frame
}

func (s *Standard) Handle(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	f := Frame{
		Timestamp: time.Now(),
		Response:  resp,
	}

	if resp.Body != nil {
		data, _ := io.ReadAll(resp.Body)
		f.Body = unsafe.String(unsafe.SliceData(data), len(data))
		resp.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	s.Channel <- f
	return resp
}

func (s *Standard) Log(ctx context.Context) {
	for {
		select {
		case f := <-s.Channel:
			fmt.Fprintf(s.Writer, "[%s] %s\t→ %s | %s\n",
				f.Timestamp.Format(time.TimeOnly),
				f.Request.Method, f.Request.URL, f.Status,
			)
			if len(f.Body) > 0 {
				fmt.Fprintln(s.Writer, f.Body)
			}
		case <-ctx.Done():
			return
		}
	}
}
