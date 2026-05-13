package log

import (
	"context"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
)

type Logger interface {
	goproxy.RespHandler
	Log(context.Context)
}

type Frame struct {
	*http.Response

	Timestamp time.Time
	Body      []byte
}
