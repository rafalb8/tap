package main

import (
	"context"
	"errors"
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

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var l log.Logger
	switch {
	case config.Json:
		l = &log.Json{Writer: os.Stdout, Channel: make(chan log.Frame, 8)}
	case config.Simple:
		l = &log.Simple{Writer: os.Stdout, Channel: make(chan log.Frame, 8)}
	default:
		l = &log.Standard{Writer: os.Stdout, Channel: make(chan log.Frame, 8)}
	}
	go l.Log(ctx)

	srv, err := StartServer(l)
	if err != nil {
		panic(err)
	}

	cmd := exec.CommandContext(ctx, config.Name, config.Args...)
	cmd.Env = []string{
		"SSL_CERT_FILE=" + config.Cert,
		"ALL_PROXY=" + srv.Addr,
		"NO_PROXY=localhost,127.0.0.1",
	}
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	err = srv.Shutdown(ctx)
	if err != nil {
		panic(err)
	}
}
