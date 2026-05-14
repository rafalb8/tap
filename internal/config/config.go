package config

import (
	"flag"
	"os"
	"path/filepath"
	"slices"

	"github.com/elazarl/goproxy"
)

// Flags
var (
	Json    bool
	Simple  bool
	Verbose bool
	Cert    string
)

var (
	Name string
	Args []string
)

func init() {
	flag.BoolVar(&Json, "json", false, "json mode")
	flag.BoolVar(&Simple, "simple", false, "simple mode")
	flag.BoolVar(&Verbose, "v", false, "verbose info")
	flag.StringVar(&Cert, "cert", filepath.Join(os.TempDir(), "tap-ca.pem"), "proxy cert")

	flags := os.Args[1:]
	split := slices.Index(os.Args, "--")
	if split != -1 {
		flags = os.Args[1:split]
		Name = os.Args[split+1]
		Args = os.Args[split+2:]
	}

	err := flag.CommandLine.Parse(flags)
	if err != nil {
		panic(err)
	}

	_, err = os.Stat(Cert)
	if os.IsNotExist(err) {
		err = os.WriteFile(Cert, goproxy.CA_CERT, 0o644)
		if err != nil {
			panic(err)
		}
	}
}
