package log

import (
	"encoding/json"
	"iter"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/rafalb8/tap/internal/flag"
)

type Json struct {
	*json.Encoder
}

type frame struct {
	Timestamp time.Time
	Latency   time.Duration `json:",omitempty"`

	Method, URL string
	Status      int `json:",omitempty"`

	Headers http.Header
}

func (j *Json) Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	ts := time.Now()
	ctx.UserData = ts

	if req == nil {
		j.Encode("Request is nil")
		return nil, nil
	}

	it := maps.All(req.Header)
	if !flag.Verbose {
		it = filterHeaders(it)
	}

	j.Encode(frame{
		Timestamp: ts,
		Method:    req.Method,
		URL:       req.URL.String(),
		Headers:   maps.Collect(it),
	})

	return req, nil
}

func (j *Json) Response(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	ts := time.Now()
	if resp == nil {
		j.Encode("Response is nil")
		return nil
	}

	it := maps.All(resp.Header)
	if !flag.Verbose {
		it = filterHeaders(it)
	}

	reqts, _ := ctx.UserData.(time.Time)
	j.Encode(frame{
		Timestamp: ts,
		Latency:   ts.Sub(reqts),
		Method:    resp.Request.Method,
		URL:       resp.Request.URL.String(),
		Status:    resp.StatusCode,
		Headers:   maps.Collect(it),
	})

	return resp
}

func filterHeaders(it iter.Seq2[string, []string]) iter.Seq2[string, []string] {
	return func(yield func(string, []string) bool) {
		for k, v := range it {
			if strings.HasPrefix(k, "X-") {
				goto yield
			}

			switch k {
			case "Content-Type", "Set-Cookie", "Location":
				goto yield
			default:
				continue
			}

		yield:
			if !yield(k, v) {
				return
			}
		}
	}
}
