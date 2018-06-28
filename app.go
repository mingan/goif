package goif

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/adammck/venv"
	"github.com/spf13/afero"
)

const (
	// EnvPrefix is the name of the ENV variable containing the grouped prefix
	EnvPrefix = "GOIF_PREFIX"
)

type App struct {
	fs  afero.Fs
	err io.Writer

	prefix        string
	excludedPaths []string
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
	app.prefix = prefix
	app.excludedPaths = strings.Split(exclude, ",")

	app.traverse("./")
}

func (app *App) traverse(dirname string) {
	files, err := afero.ReadDir(app.fs, dirname)
	if err != nil {
		app.error("reading directory", err)
		return
	}

	for _, file := range files {
		path := dirname + file.Name()
		if app.isExcluded(path) {
			continue
		}
		if file.IsDir() {
			app.traverse(path + "/")
			continue
		}
		if err := app.formatFile(path); err != nil {
			app.error(file, err)
		}
	}
}

func (app *App) isExcluded(path string) bool {
	path = strings.TrimPrefix(path, "./")
	for _, e := range app.excludedPaths {
		if e == path {
			return true
		}
	}

	return false
}

func (app *App) formatFile(path string) error {
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

	formatter := NewFormatter(app.prefix)
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
