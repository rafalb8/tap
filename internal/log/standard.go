package log

import (
	"bytes"
	"encoding/json"
	"io"
	"iter"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/rafalb8/tap/internal/config"
)

type Standard struct {
	Writer io.Writer
}

func (s *Standard) Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	panic("TODO")
}

func (s *Standard) Response(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if resp == nil {
		panic("Response is nil")
	}

	buf := &bytes.Buffer{}
	s.writeResponse(buf, resp)
	s.writeHeaders(buf, resp)
	s.writeBody(buf, resp)
	buf.WriteTo(s.Writer)

	return resp
}

func (s *Standard) writeResponse(buf *bytes.Buffer, resp *http.Response) {
	// [10:31:05] POST → https://api.example.com/v1/auth | 200 OK
	buf.WriteByte('[')
	buf.WriteString(time.Now().Format(time.TimeOnly))
	buf.WriteString("] ")
	buf.WriteString(resp.Request.Method)
	buf.WriteString("\t→ ")
	buf.WriteString(resp.Request.URL.String())
	buf.WriteString(" | ")
	buf.WriteString(resp.Status)
	buf.WriteByte('\n')
}

func (s *Standard) writeHeaders(buf *bytes.Buffer, resp *http.Response) {
	if resp.Header == nil {
		return
	}

	it := maps.Keys(resp.Header)
	if !config.Verbose {
		it = s.filterHeaders(it)
	}
	keys := slices.Collect(it)
	slices.Sort(keys)
	for _, key := range keys {
		buf.WriteString(key)
		buf.WriteString(": ")
		for _, v := range resp.Header[key] {
			buf.WriteString(v)
			buf.WriteByte(' ')
		}
		buf.WriteByte('\n')
	}
}

func (s *Standard) writeBody(buf *bytes.Buffer, resp *http.Response) {
	if resp.Body == nil {
		return
	}

	data, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(data))

	ct := resp.Header.Get("Content-Type")
	ct, _, _ = strings.Cut(ct, ";")
	typ, subTyp, _ := strings.Cut(ct, "/")
	switch typ {
	case "image", "audio", "video":
		buf.WriteString("[Binary data: ")
		buf.WriteString(subTyp)
		buf.WriteByte(' ')
		buf.WriteString(typ)
		buf.WriteString(", ")
		WriteBytes(buf, len(data))
		buf.WriteString("]\n")
		return
	case "multipart":
		buf.WriteString("[Binary data: ")
		WriteBytes(buf, len(data))
		buf.WriteString("]\n")
		return
	}

	switch subTyp {
	case "json":
		json.Indent(buf, data, "", "  ")
	default:
		buf.Write(data)
		buf.WriteByte('\n')
	}
}

func (*Standard) filterHeaders(it iter.Seq[string]) iter.Seq[string] {
	return func(yield func(string) bool) {
		for k := range it {
			switch {
			case k == "Content-Type":
			case k == "Set-Cookie":
			case strings.HasPrefix(k, "X-"):
			default:
				continue
			}
			if !yield(k) {
				return
			}
		}
	}
}

func WriteBytes(buf *bytes.Buffer, i int) {
	const unit = 1000

	if i < unit {
		buf.Write(strconv.AppendInt(buf.AvailableBuffer(), int64(i), 10))
		buf.WriteString(" B")
		return
	}

	div, exp := unit, 0
	for n := i / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	buf.Write(strconv.AppendFloat(buf.AvailableBuffer(), float64(i)/float64(div), 'f', 2, 32))
	buf.WriteByte(' ')
	buf.WriteByte("kMGTPE"[exp])
}
