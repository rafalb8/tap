package flag

import (
	"crypto/tls"
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

	Output string

	CertFile string
)

var (
	Name string
	Args []string
)

func init() {
	defCertDir := filepath.Join(os.TempDir(), "tap")

	pflag.BoolVarP(&Json, "json", "j", false, "json mode")
	pflag.BoolVarP(&Simple, "simple", "s", false, "simple mode")
	pflag.BoolVarP(&Verbose, "verbose", "v", false, "print all headers (json/default mode only)")
	pflag.StringVarP(&Output, "output", "o", "", "output path; enables std outputs")
	certDir := pflag.String("cert-dir", defCertDir, "proxy cert directory")

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

	CertFile = filepath.Join(*certDir, "ca.crt")

	if *certDir == defCertDir {
		saveCert(*certDir)
	} else {
		loadCert(*certDir)
	}
}

func saveCert(dir string) {
	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		panic(err)
	}

	_, err = os.Stat(CertFile)
	if os.IsNotExist(err) {
		err = os.WriteFile(CertFile, goproxy.CA_CERT, 0o644)
		if err != nil {
			panic(err)
		}
	}
}

func loadCert(dir string) {
	_, err := os.Stat(CertFile)
	if err != nil {
		panic("loadCert: certFile: " + err.Error())
	}

	keyFile := filepath.Join(dir, "ca.key")
	_, err = os.Stat(keyFile)
	if err != nil {
		panic("loadCert: keyFile: " + err.Error())
	}

	cert, err := tls.LoadX509KeyPair(CertFile, keyFile)
	if err != nil {
		panic("loadCert: failed: " + err.Error())
	}

	goproxy.GoproxyCa = cert
}
