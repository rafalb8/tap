package log

import (
	"bytes"
	"context"
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
	buf := &bytes.Buffer{}
	defer buf.WriteTo(s.Writer)

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
		s.writeHeaders(buf, f)
	}

	if len(f.Body) > 0 {
		ct := f.Header.Get("Content-Type")
		ct, _, _ = strings.Cut(ct, ";")
		typ, subTyp, _ := strings.Cut(ct, "/")
		switch typ {
		case "image", "audio", "video":
			buf.WriteString("[Binary data: ")
			buf.WriteString(subTyp)
			buf.WriteByte(' ')
			buf.WriteString(typ)
			buf.WriteString(", ")
			WriteBytes(buf, len(f.Body))
			buf.WriteString("]\n")
			return
		case "multipart":
			buf.WriteString("[Binary data: ")
			WriteBytes(buf, len(f.Body))
			buf.WriteString("]\n")
			return
		}

		switch subTyp {
		case "json":
			json.Indent(buf, f.Body, "", "  ")
		default:
			buf.Write(f.Body)
			buf.WriteByte('\n')
		}
	}
}

func (s *Standard) writeHeaders(buf *bytes.Buffer, f Frame) {
	it := maps.Keys(f.Header)
	if !config.Verbose {
		it = s.filterHeaders(it)
	}
	keys := slices.Collect(it)
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
