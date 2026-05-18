package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/rafalb8/tap/internal/executor"
	"github.com/rafalb8/tap/internal/flag"
	"github.com/rafalb8/tap/internal/log"
	"github.com/rafalb8/tap/internal/server"
)

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

	srv, err := server.New(log.New(w))
	if err != nil {
		panic(err)
	}

	err = executor.Run(ctx, srv.Addr)
	if err != nil {
		panic(err)
	}

	err = srv.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
}
