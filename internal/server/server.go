package server

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/rafalb8/tap/internal/log"
)

func New(middleware log.Middleware) (*http.Server, error) {
	prxy := goproxy.NewProxyHttpServer()
	prxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	prxy.OnRequest().DoFunc(middleware.Request)
	prxy.OnResponse().DoFunc(middleware.Response)

	// Find free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	srv := &http.Server{Handler: prxy, Addr: l.Addr().String()}
	idx := strings.LastIndex(srv.Addr, ":")
	if idx < 0 {
		return nil, errors.New("missing port")
	}
	srv.Addr = "127.0.0.1" + srv.Addr[idx:]

	go func() {
		err = srv.Serve(l)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return srv, nil
}
