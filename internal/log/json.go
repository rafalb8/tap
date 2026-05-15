package log

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/rafalb8/tap/internal/flag"
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

	it := maps.All(req.Header)
	if !flag.Verbose {
		it = filterHeaders(it)
	}

	// encode request
	enc := json.NewEncoder(j.Writer)
	enc.Encode(frame{
		Timestamp: ts,
		Method:    req.Method,
		URL:       req.URL.String(),
		Headers:   maps.Collect(it),
	})

	// make request
	resp, err := ctx.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(j.Writer, "Response error: %v\n", err)
		return req, nil
	}
	end := time.Now()

	it = maps.All(resp.Header)
	if !flag.Verbose {
		it = filterHeaders(it)
	}

	// encode response
	enc.Encode(frame{
		Timestamp: end,
		Latency:   end.Sub(ts),
		Method:    req.Method,
		URL:       req.URL.String(),
		Status:    resp.StatusCode,
		Headers:   maps.Collect(it),
	})

	return req, resp
}

func filterHeaders(it iter.Seq2[string, []string]) iter.Seq2[string, []string] {
	return func(yield func(string, []string) bool) {
		for k, v := range it {
			switch {
			case k == "Content-Type":
			case k == "Set-Cookie":
			case strings.HasPrefix(k, "X-"):
			default:
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}
