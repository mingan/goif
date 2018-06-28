package goif

import (
	"github.com/spf13/afero"
	"github.com/adammck/venv"
	"io"
	"fmt"
	"io/ioutil"
	"strings"
	"os"
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

func (app *App) Run(prefix, exclude string, env venv.Env) {
	if prefix == "" {
		prefix = env.Getenv(EnvPrefix)
	}

	app.traverse(prefix, "", strings.Split(exclude, ","))
}

func (app *App) traverse(prefix, dirname string, excludedPaths []string) {
	files, err := afero.ReadDir(app.fs, dirname)
	if err != nil {
		app.error("reading directory", err)
		return
	}

	for _, file := range files {
		path := dirname + file.Name()
		if isExcluded(path, excludedPaths) {
			continue
		}
		if file.IsDir() {
			app.traverse(prefix, path+"/", excludedPaths)
			continue
		}
		if err := app.formatFile(prefix, path); err != nil {
			app.error(file, err)
		}
	}
}

func isExcluded(path string, excludedPaths []string) bool {
	for _, e := range excludedPaths {
		if e == path {
			return true
		}
	}

	return false
}

func (app *App) formatFile(prefix, path string) error {
	file, err := app.fs.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	safePath := strings.Replace(path, string(os.PathSeparator), "_", -1)
	temp, err := ioutil.TempFile("", safePath)
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

func (app *App) error(args ... interface{}) {
	fmt.Fprintln(app.err, args...)
}
