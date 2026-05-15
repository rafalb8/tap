package log

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
)

type Json struct {
	Writer io.Writer
}

type frame struct {
	Timestamp time.Time
	Latency   time.Duration `json:",omitempty"`

	Method, URL string
	Status      int `json:",omitempty"`

	Headers http.Header
}

func (j *Json) Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	if req == nil {
		fmt.Fprintln(j.Writer, "Request is nil")
		return nil, nil
	}
	ts := time.Now()

	// encode request
	enc := json.NewEncoder(j.Writer)
	enc.Encode(frame{
		Timestamp: ts,
		Method:    req.Method,
		URL:       req.URL.String(),
		Headers:   req.Header,
	})

	// make request
	resp, err := ctx.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(j.Writer, "Response error: %v\n", err)
		return req, nil
	}
	end := time.Now()

	// encode response
	enc.Encode(frame{
		Timestamp: end,
		Latency:   end.Sub(ts),
		Method:    req.Method,
		URL:       req.URL.String(),
		Status:    resp.StatusCode,
		Headers:   resp.Header,
	})

	return req, resp
}
