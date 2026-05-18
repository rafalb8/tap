package executor

import (
	"context"
	"os"
	"os/exec"

	"github.com/rafalb8/tap/internal/flag"
)

func Run(ctx context.Context, addr string) error {
	cmd := exec.CommandContext(ctx, flag.Name, flag.Args...)
	cmd.Env = append(os.Environ(),
		"SSL_CERT_FILE="+flag.CertFile,
		"ALL_PROXY="+addr,
		"NO_PROXY=localhost,127.0.0.1",
	)

	if flag.Output != "" {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}
