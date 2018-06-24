package goif

import (
	"testing"
	"github.com/spf13/afero"
	"github.com/adammck/venv"
	"bytes"
	"github.com/andreyvit/diff"
)

func TestApp_Run(t *testing.T) {
	t.Run("single file, without flag, without ENV, doesn't change the content", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		NewApp(fs, &stderr).Run("", venv.Mock())

		content, err := readFile(fs, "main.go")
		if err != nil {
			t.Fatal(err)
		}

		if content != singleFileNoPrefix {
			t.Error(diff.LineDiff(singleFileNoPrefix, content))
		}
	})

	t.Run("single file, with flag, without ENV", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		NewApp(fs, &stderr).Run("acme.com", venv.Mock())

		content, err := readFile(fs, "main.go")
		if err != nil {
			t.Fatal(err)
		}

		if content != singleFileAcme {
			t.Error(diff.LineDiff(singleFileAcme, content))
		}
	})

	t.Run("single file, without flag, with ENV", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		env := venv.Mock()
		env.Setenv("GOIF_PREFIX", "acme.com")
		NewApp(fs, &stderr).Run("", env)

		content, err := readFile(fs, "main.go")
		if err != nil {
			t.Fatal(err)
		}

		if content != singleFileAcme {
			t.Error(diff.LineDiff(singleFileAcme, content))
		}
	})

	t.Run("single file, with flag and with ENV, flag takes precedence", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		env := venv.Mock()
		env.Setenv("GOIF_PREFIX", "foobar.com")
		NewApp(fs, &stderr).Run("acme.com", env)

		content, err := readFile(fs, "main.go")
		if err != nil {
			t.Fatal(err)
		}

		if content != singleFileAcme {
			t.Error(diff.LineDiff(singleFileAcme, content))
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
	fs := afero.NewMemMapFs()
	file, _ := fs.Create("main.go")
	file.WriteString(singleFileSource)
	file.Close()
	return fs
}

const (
	singleFileSource = `
package main

import (
	"github.com/some/package"
	"log"
	"foobar.com/useful/package"
	"fmt"
	"acme.com/awesome/package"
)

func main() {
	fmt.Println("Hello world")
}
`
	singleFileNoPrefix = `
package main

import (
	"fmt"
	"log"

	"acme.com/awesome/package"
	"foobar.com/useful/package"
	"github.com/some/package"
)

func main() {
	fmt.Println("Hello world")
}
`
	singleFileAcme = `
package main

import (
	"fmt"
	"log"

	"acme.com/awesome/package"

	"foobar.com/useful/package"
	"github.com/some/package"
)

func main() {
	fmt.Println("Hello world")
}
`
)

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
