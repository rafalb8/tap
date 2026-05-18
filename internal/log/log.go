package log

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/rafalb8/tap/internal/flag"
)

type Middleware interface {
	Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response)
	Response(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response
}

func New(w io.Writer) Middleware {
	switch {
	case flag.Json:
		return &Json{Encoder: json.NewEncoder(w)}
	case flag.Simple:
		return &Simple{Writer: bufio.NewWriter(w)}
	default:
		return &Pretty{Writer: bufio.NewWriter(w)}
	}
}
