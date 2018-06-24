package main

import (
	"github.com/spf13/afero"
	"github.com/adammck/venv"
	"os"
	"github.com/mingan/goif"
	"flag"
)

var (
	flagPrefix = flag.String("prefix", "", "Prefix ") 
)

func main() {
	// get prefix from an env var or from an arg
	goif.NewApp(afero.NewOsFs(), os.Stderr).Run(*flagPrefix, venv.OS())
}
