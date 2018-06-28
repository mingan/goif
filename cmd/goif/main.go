package main

import (
	"flag"
	"os"

	"github.com/adammck/venv"
	"github.com/mingan/goif"
	"github.com/spf13/afero"
)

var (
	flagPrefix  = flag.String("prefix", "", "Prefix ")
	flagExclude = flag.String("exclude", "vendor", "Comma-separated list of glob patterns to exclude")
)

func main() {
	// get prefix from an env var or from an arg
	goif.NewApp(afero.NewOsFs(), os.Stderr).
		Run(*flagPrefix, *flagExclude, venv.OS())
}
