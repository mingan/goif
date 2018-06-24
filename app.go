package goif

import (
	"github.com/spf13/afero"
	"github.com/adammck/venv"
	"io"
	"fmt"
	"io/ioutil"
)

const (
	// EnvPrefix is the name of the ENV variable containing the grouped prefix
	EnvPrefix = "GOIF_PREFIX"
)

type App struct {
	fs  afero.Fs
	err io.Writer
}

func NewApp(fs afero.Fs, err io.Writer) *App {
	return &App{
		fs:  fs,
		err: err,
	}
}

func (app *App) Run(prefix string, env venv.Env) {
	if prefix == "" {
		prefix = env.Getenv(EnvPrefix)
	}
	
	// given glob path
	paths, err := afero.Glob(app.fs, "*.go")
	if err != nil {
		fmt.Fprint(app.err, err)
		return
	}

	// traverse all go files
	// process each one of them, in parallel?
	for _, path := range paths {
		if err := app.formatFile(prefix, path); err != nil {
			// print errors to stderr
			fmt.Fprint(app.err, err)
		}
	}
}

func (app *App) formatFile(prefix, path string) error {
	file, err := app.fs.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	temp, err := ioutil.TempFile("", path)
	if err != nil {
		return err
	}
	defer temp.Close()

	formatter := NewFormatter(prefix)
	// format file to temp file
	if err := formatter.Format(file, temp); err != nil {
		// if an error occurs, scrap the temp file, return error
		return err
	}

	// close file for reads
	if err := file.Close(); err != nil {
		return err
	}

	// reopen for write, scrapping it
	file, err = app.fs.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// reset cursor to the beginning of the file
	if _, err := temp.Seek(0, 0); err != nil {
		return err
	}

	// replace the file by the temp
	if _, err := io.Copy(file, temp); err != nil {
		return err
	}

	return nil
}
