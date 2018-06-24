package goif

import (
	"testing"
	"github.com/spf13/afero"
	"github.com/adammck/venv"
	"bytes"
	"github.com/andreyvit/diff"
)

func TestApp_Run(t *testing.T) {
	t.Run("single file", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		NewApp(fs, &stderr).Run(venv.Mock())

		content, err := readFile(fs, "main.go")
		if err != nil {
			t.Fatal(err)
		}

		expected := `
package main

import (
	"fmt"
	"log"

	"acme.com/awesome/package"
	"github.com/some/package"
)

func main() {
	fmt.Println("Hello world")
}
`
		if content != expected {
			t.Error(diff.LineDiff(expected, content))
		}
	})

	// arg
	// env var
	// neither
	// error
	// recursiveness
	// different location
}

func prepareSingleTestFile() afero.Fs {
	input := `
package main

import (
	"github.com/some/package"
	"log"
	"fmt"
	"acme.com/awesome/package"
)

func main() {
	fmt.Println("Hello world")
}
`
	fs := afero.NewMemMapFs()
	file, _ := fs.Create("main.go")
	file.WriteString(input)
	file.Close()
	return fs
}

// readFile basically duplicates ioutil.ReadFile except it returns a string
func readFile(fs afero.Fs, filename string) (string, error) {
	file, err := fs.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var n int64
	if fi, err := file.Stat(); err == nil {
		// Don't preallocate a huge buffer, just in case.
		if size := fi.Size(); size < 1e9 {
			n = size
		}
	}
	buf := bytes.NewBuffer(make([]byte, 0, n+bytes.MinRead))
	if _, err := buf.ReadFrom(file); err != nil {
		return "", err
	}
	return buf.String(), nil
}
