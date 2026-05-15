package log

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type Pretty struct {
	Writer io.Writer
}

func (p *Pretty) Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	if req == nil {
		fmt.Fprintln(p.Writer, "Request is nil")
		return nil, nil
	}
	ts := time.Now()

	buf := &bytes.Buffer{}
	writeRequest(buf, req)
	p.writeHeaders(buf, req.Header)
	p.writeBody(buf, req.Header.Get("Content-Type"), &req.Body)
	buf.WriteTo(p.Writer)

	resp, err := ctx.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(p.Writer, "Response error: %v\n", err)
		return req, nil
	}
	latency := time.Since(ts)

	buf.Reset()
	writeResponse(buf, resp, latency)
	p.writeHeaders(buf, resp.Header)
	p.writeBody(buf, resp.Header.Get("Content-Type"), &resp.Body)
	buf.WriteTo(p.Writer)

	return req, resp
}

func (p *Pretty) writeHeaders(buf *bytes.Buffer, header http.Header) {
	if len(header) == 0 {
		return
	}

	it := maps.Keys(header)
	if !config.Verbose {
		it = p.filterHeaders(it)
	}
	keys := slices.Collect(it)
	slices.Sort(keys)
	for _, key := range keys {
		buf.WriteString(key)
		buf.WriteString(": ")
		for _, v := range header[key] {
			buf.WriteString(v)
			buf.WriteByte(' ')
		}
		buf.WriteByte('\n')
	}
}

func (p *Pretty) writeBody(buf *bytes.Buffer, ct string, body *io.ReadCloser) {
	data, _ := io.ReadAll(*body)
	*body = io.NopCloser(bytes.NewBuffer(data))

	if len(data) == 0 {
		return
	}

	ct, _, _ = strings.Cut(ct, ";")
	typ, subTyp, _ := strings.Cut(ct, "/")
	switch typ {
	case "image", "audio", "video":
		buf.WriteString("[Binary data: ")
		buf.WriteString(subTyp)
		buf.WriteByte(' ')
		buf.WriteString(typ)
		buf.WriteString(", ")
		formatBytes(buf, len(data))
		buf.WriteString("]\n")
		return
	case "multipart":
		buf.WriteString("[Binary data: ")
		formatBytes(buf, len(data))
		buf.WriteString("]\n")
		return
	}

	switch subTyp {
	case "json":
		json.Indent(buf, data, "", "  ")
		buf.WriteByte('\n')
	default:
		buf.Write(data)
		buf.WriteByte('\n')
	}
}

func (*Pretty) filterHeaders(it iter.Seq[string]) iter.Seq[string] {
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

func formatBytes(buf *bytes.Buffer, i int) {
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
