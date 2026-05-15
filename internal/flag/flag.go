package flag

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/elazarl/goproxy"
	"github.com/spf13/pflag"
)

// Flags
var (
	Json    bool
	Simple  bool
	Verbose bool

	Cert   string
	Output string
)

var (
	Name string
	Args []string
)

func init() {
	pflag.BoolVarP(&Json, "json", "j", false, "json mode")
	pflag.BoolVarP(&Simple, "simple", "s", false, "simple mode")
	pflag.BoolVarP(&Verbose, "verbose", "v", false, "print all headers (json/default mode only)")
	pflag.StringVar(&Cert, "cert", filepath.Join(os.TempDir(), "tap-ca.pem"), "proxy cert")
	pflag.StringVarP(&Output, "output", "o", "", "output path; enables std outputs")

	flags := os.Args[1:]
	split := slices.Index(os.Args, "--")
	if split != -1 {
		flags = os.Args[1:split]
		Name = os.Args[split+1]
		Args = os.Args[split+2:]
	}

	err := pflag.CommandLine.Parse(flags)
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
