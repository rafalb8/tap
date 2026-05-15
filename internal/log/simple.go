package log

import (
	"bufio"
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

	// print request
	buf := bufio.NewWriter(s.Writer)
	writeRequest(buf, req, ts)
	buf.Flush()

	// make request
	resp, err := ctx.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(s.Writer, "Response error: %v\n", err)
		return req, nil
	}
	end := time.Now()

	// print response
	writeResponse(buf, resp, end, end.Sub(ts))
	buf.Flush()

	return req, resp
}

type buffer interface {
	WriteByte(c byte) error
	WriteString(s string) (int, error)
}

func writeRequest(buf buffer, req *http.Request, ts time.Time) {
	// [10:31:05] POST → https://api.example.com/v1/auth
	buf.WriteByte('[')
	buf.WriteString(ts.Format(time.TimeOnly))
	buf.WriteString("] ")
	buf.WriteString(req.Method)
	buf.WriteString("\t→ ")
	buf.WriteString(req.URL.String())
	buf.WriteByte('\n')
}

func writeResponse(buf buffer, resp *http.Response, ts time.Time, latency time.Duration) {
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
