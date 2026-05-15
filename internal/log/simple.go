package log

import (
	"bytes"
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
	if req == nil {
		fmt.Fprintln(s.Writer, "Request is nil")
		return nil, nil
	}
	ts := time.Now()

	buf := &bytes.Buffer{}
	writeRequest(buf, req)
	buf.WriteTo(s.Writer)

	resp, err := ctx.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(s.Writer, "Response error: %v\n", err)
		return req, nil
	}
	latency := time.Since(ts)

	buf.Reset()
	writeResponse(buf, resp, latency)
	buf.WriteTo(s.Writer)

	return req, resp
}

func writeRequest(buf *bytes.Buffer, req *http.Request) {
	// [10:31:05] POST → https://api.example.com/v1/auth
	buf.WriteByte('[')
	buf.WriteString(time.Now().Format(time.TimeOnly))
	buf.WriteString("] ")
	buf.WriteString(req.Method)
	buf.WriteString("\t→ ")
	buf.WriteString(req.URL.String())
	buf.WriteByte('\n')
}

func writeResponse(buf *bytes.Buffer, resp *http.Response, latency time.Duration) {
	// [10:31:05] POST ← https://api.example.com/v1/auth | 200 OK (50ms)
	buf.WriteByte('[')
	buf.WriteString(time.Now().Format(time.TimeOnly))
	buf.WriteString("] ")
	buf.WriteString(resp.Request.Method)
	buf.WriteString("\t← ")
	buf.WriteString(resp.Request.URL.String())
	buf.WriteString(" | ")
	buf.WriteString(resp.Status)
	buf.WriteString(" (")
	buf.WriteString(latency.Truncate(time.Millisecond).String())
	buf.WriteString(")\n")
}
