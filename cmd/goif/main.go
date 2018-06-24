package main

import (
	"github.com/spf13/afero"
	"github.com/adammck/venv"
	"os"
	"github.com/mingan/goif"
)

func main() {
	// get prefix from an env var or from an arg
	goif.NewApp(afero.NewOsFs(), os.Stderr).Run(venv.OS())
}
