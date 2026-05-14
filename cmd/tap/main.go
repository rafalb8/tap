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

	"github.com/elazarl/goproxy"
	"github.com/rafalb8/tap/internal/config"
	"github.com/rafalb8/tap/internal/log"
)

func StartServer(l log.Logger) (*http.Server, error) {
	prxy := goproxy.NewProxyHttpServer()
	prxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	prxy.OnResponse().Do(l)

	// Find free port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	srv := &http.Server{Handler: prxy, Addr: ln.Addr().String()}
	go func() {
		err = srv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	return srv, nil
}

func NewLogger(w io.Writer) log.Logger {
	c := make(chan log.Frame, 8)
	switch {
	case config.Json:
		return &log.Json{Writer: w, Channel: c}
	case config.Simple:
		return &log.Simple{Writer: w, Channel: c}
	default:
		return &log.Standard{Writer: w, Channel: c}
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	w := os.Stdout
	if config.Output != "" {
		f, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		w = f
	}

	l := NewLogger(w)
	go l.Log(ctx)

	srv, err := StartServer(l)
	if err != nil {
		panic(err)
	}

	cmd := exec.CommandContext(ctx, config.Name, config.Args...)
	cmd.Env = append(os.Environ(),
		"SSL_CERT_FILE="+config.Cert,
		"ALL_PROXY="+srv.Addr,
		"NO_PROXY=localhost,127.0.0.1",
	)
	if config.Output != "" {
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
