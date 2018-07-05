package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adammck/venv"
	"github.com/mingan/goif"
	"github.com/spf13/afero"
)

func flags() (string, string, bool) {
	var prefix, exclude string
	var help bool

	flag.StringVar(&prefix, "prefix", "", "Prefix to be grouped separately, e.g. github.com/mycompany")
	flag.StringVar(&exclude, "exclude", "vendor", "Comma-separated list of glob patterns to exclude")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Parse()

	return prefix, exclude, help
}

func main() {
	prefix, exclude, help := flags()

	if help {
		printHelp()
		return
	}
	
	goif.NewApp(afero.NewOsFs(), os.Stderr).
		Run(prefix, exclude, venv.OS())
}

func printHelp() {
	fmt.Fprintln(os.Stderr, "Usage: goif --prefix=github.com/mycompany", "\n\nOptions:")
	flag.PrintDefaults()
}
