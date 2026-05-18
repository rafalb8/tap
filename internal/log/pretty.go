package log

import (
	"bufio"
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
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/rafalb8/tap/internal/flag"
)

var bytePool = sync.Pool{New: func() any { return &bytes.Buffer{} }}

type Pretty struct {
	Writer *bufio.Writer
}

func (p *Pretty) Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	ts := time.Now()
	ctx.UserData = ts

	if req == nil {
		fmt.Fprintln(p.Writer, "Request is nil")
		return nil, nil
	}

	writeRequest(p.Writer, req, ts)
	writeHeaders(p.Writer, req.Header)
	writeBody(p.Writer, req.Header.Get("Content-Type"), &req.Body)
	p.Writer.Flush()

	return req, nil
}

func (p *Pretty) Response(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	ts := time.Now()
	if resp == nil {
		fmt.Fprintln(p.Writer, "Response is nil")
		return nil
	}

	reqts, _ := ctx.UserData.(time.Time)
	writeResponse(p.Writer, resp, ts, ts.Sub(reqts))
	writeHeaders(p.Writer, resp.Header)
	writeBody(p.Writer, resp.Header.Get("Content-Type"), &resp.Body)
	p.Writer.Flush()

	return resp
}

func writeHeaders(buf *bufio.Writer, header http.Header) {
	if len(header) == 0 {
		return
	}

	it := maps.Keys(header)
	if !flag.Verbose {
		it = seq2one(filterHeaders(maps.All(header)))
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

func writeBody(buf *bufio.Writer, ct string, body *io.ReadCloser) {
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
		dst := bytePool.Get().(*bytes.Buffer)
		defer bytePool.Put(dst)
		dst.Reset()

		json.Indent(dst, data, "", "  ")
		dst.WriteTo(buf)
		buf.WriteByte('\n')
	default:
		buf.Write(data)
		buf.WriteByte('\n')
	}
}

func seq2one[K, V any](it iter.Seq2[K, V]) iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range it {
			if !yield(k) {
				return
			}
		}
	}
}

func formatBytes(buf *bufio.Writer, i int) {
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
