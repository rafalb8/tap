package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/rafalb8/tap/internal/flag"
	"github.com/rafalb8/tap/internal/log"
)

type Logger interface {
	Request(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response)
}

func NewLogger(w io.Writer) Logger {
	switch {
	case flag.Json:
		return &log.Json{Writer: w}
	case flag.Simple:
		return &log.Simple{Writer: w}
	default:
		return &log.Pretty{Writer: w}
	}
}

func StartServer(middleware Logger) (*http.Server, error) {
	prxy := goproxy.NewProxyHttpServer()
	prxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	prxy.OnRequest().DoFunc(middleware.Request)

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

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	w := os.Stdout
	if flag.Output != "" {
		f, err := os.OpenFile(flag.Output, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		w = f
	}

	srv, err := StartServer(NewLogger(w))
	if err != nil {
		panic(err)
	}

	cmd := exec.CommandContext(ctx, flag.Name, flag.Args...)
	cmd.Env = append(os.Environ(),
		"SSL_CERT_FILE="+flag.CertFile,
		"ALL_PROXY="+srv.Addr,
		"NO_PROXY=localhost,127.0.0.1",
	)

	if flag.Output != "" {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	err = srv.Shutdown(ctx)
	if err != nil {
		panic(err)
	}
}
