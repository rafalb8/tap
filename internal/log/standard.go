package log

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"maps"
	"net/http"
	"slices"
	"time"

	"github.com/elazarl/goproxy"
)

type Standard struct {
	Writer  io.Writer
	Channel chan Frame
}

func (s *Standard) Handle(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if resp == nil {
		panic("Response is nil")
	}

	f := Frame{
		Timestamp: time.Now(),
		Response:  resp,
	}

	if resp.Body != nil {
		f.Body, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(f.Body))
	}

	s.Channel <- f
	return resp
}

func (s *Standard) Log(ctx context.Context) {
	for {
		select {
		case f := <-s.Channel:
			s.Write(f)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Standard) Write(f Frame) {
	buf := bufio.NewWriter(s.Writer)

	// [10:31:05] POST → https://api.example.com/v1/auth | 200 OK
	buf.WriteByte('[')
	buf.WriteString(f.Timestamp.Format(time.TimeOnly))
	buf.WriteString("] ")
	buf.WriteString(f.Request.Method)
	buf.WriteString("\t→ ")
	buf.WriteString(f.Request.URL.String())
	buf.WriteString(" | ")
	buf.WriteString(f.Status)
	buf.WriteByte('\n')

	// Headers
	if len(f.Header) > 0 {
		keys := slices.Collect(maps.Keys(f.Header))
		slices.Sort(keys)
		for _, key := range keys {
			buf.WriteString(key)
			buf.WriteString(": ")
			for _, v := range f.Header[key] {
				buf.WriteString(v)
				buf.WriteByte(' ')
			}
			buf.WriteByte('\n')
		}
	}

	if len(f.Body) > 0 {
		buf.Write(f.Body)
		buf.WriteByte('\n')
	}

	buf.Flush()
}
