package log

import (
	"bufio"
	"fmt"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
)

type Simple struct {
	Writer *bufio.Writer
}

func (s *Simple) Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	ts := time.Now()
	ctx.UserData = ts

	if req == nil {
		fmt.Fprintln(s.Writer, "Request is nil")
		return nil, nil
	}

	writeRequest(s.Writer, req, ts)
	s.Writer.Flush()

	return req, nil
}

func (s *Simple) Response(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	ts := time.Now()
	if resp == nil {
		fmt.Fprintln(s.Writer, "Response is nil")
		return nil
	}

	reqts, _ := ctx.UserData.(time.Time)
	writeResponse(s.Writer, resp, ts, ts.Sub(reqts))
	s.Writer.Flush()

	return resp
}

func writeRequest(buf *bufio.Writer, req *http.Request, ts time.Time) {
	// [10:31:05] POST → https://api.example.com/v1/auth
	buf.WriteByte('[')
	buf.WriteString(ts.Format(time.TimeOnly))
	buf.WriteString("] ")
	buf.WriteString(req.Method)
	buf.WriteString("\t→ ")
	buf.WriteString(req.URL.String())
	buf.WriteByte('\n')
}

func writeResponse(buf *bufio.Writer, resp *http.Response, ts time.Time, latency time.Duration) {
	// [10:31:05] POST ← https://api.example.com/v1/auth | 200 OK (50ms)
	buf.WriteByte('[')
	buf.WriteString(ts.Format(time.TimeOnly))
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
