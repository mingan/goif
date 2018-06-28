package goif

import (
	"testing"
	"github.com/spf13/afero"
	"github.com/adammck/venv"
	"bytes"
	"github.com/andreyvit/diff"
	"fmt"
	"runtime/debug"
	"strings"
	"os"
	"regexp"
)

func TestApp_Run(t *testing.T) {

	t.Run("single file, without flag, without ENV, doesn't change the content", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		NewApp(fs, &stderr).Run("", "", venv.Mock())

		expectContentToMatchString(fs, singleFilePath, singleFileNoPrefix, t)
	})

	t.Run("single file, with flag, without ENV", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		NewApp(fs, &stderr).Run("acme.com", "", venv.Mock())

		expectContentToMatchString(fs, singleFilePath, singleFileAcme, t)
	})

	t.Run("single file, without flag, with ENV", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		env := venv.Mock()
		env.Setenv("GOIF_PREFIX", "acme.com")
		NewApp(fs, &stderr).Run("", "", env)

		expectContentToMatchString(fs, singleFilePath, singleFileAcme, t)
	})

	t.Run("single file, with flag and with ENV, flag takes precedence", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		env := venv.Mock()
		env.Setenv("GOIF_PREFIX", "foobar.com")
		NewApp(fs, &stderr).Run("acme.com", "", env)

		expectContentToMatchString(fs, singleFilePath, singleFileAcme, t)
	})

	t.Run("different file, the tested file is not touched", func(t *testing.T) {
		t.Parallel()

		fs := prepareSingleTestFile()
		stderr := bytes.Buffer{}
		NewApp(fs, &stderr).Run(
			"acme.com",
			fmt.Sprintf("%v,other_file.go", singleFilePath),
			venv.Mock(),
		)

		expectContentToMatchString(fs, singleFilePath, singleFileOriginal, t)
	})

	t.Run("nested structure, all non-excluded files are formatted", func(t *testing.T) {
		t.Parallel()

		fs := prepareNestedTestFiles()
		stderr := bytes.Buffer{}
		NewApp(fs, &stderr).Run("acme.com", "", venv.Mock())

		expectContentToMatchString(fs, singleFilePath, singleFileAcme, t)
		expectContentToMatchString(fs, childFilePath, childFileAcme, t)
		expectContentToMatchString(fs, grandchildFilePath, grandchildFileAcme, t)
	})
}

func expectContentToMatchString(fs afero.Fs, path, expected string, t *testing.T) {
	content, err := readFile(fs, path)
	if err != nil {
		t.Fatal(err, path)
	}

	if content != expected {
		t.Error(originalCaller(1), diff.LineDiff(expected, content), path)
	}
}

func originalCaller(depth int) string {
	if depth < 1 {
		depth = 1
	}
	// goroutine, debug.Stack + self lines + depth
	depth = 1 + (depth+2)*2

	lines := strings.Split(string(debug.Stack()), "\n")

	return fmt.Sprintf(
		"%s (%s)",
		caller(lines, depth),
		callSite(lines, depth),
	)
}
func caller(lines []string, depth int) string {
	m := callerRegexp.FindStringSubmatch(lines[depth])
	if len(m) >= 2 {
		return m[1]
	}

	return "?"
}

func callSite(lines []string, depth int) string {
	m := callSiteRegexp.FindStringSubmatch(lines[depth+1])
	if len(m) >= 2 {
		cwd, _ := os.Getwd()
		return strings.TrimPrefix(m[1], cwd+"/")
	}

	return "?"
}

func prepareSingleTestFile() afero.Fs {
	fs := afero.NewMemMapFs()
	writeTestFile(fs, singleFilePath, singleFileOriginal)
	return fs
}
func writeTestFile(fs afero.Fs, path, content string) {
	file, _ := fs.Create(path)
	file.WriteString(content)
	file.Close()
}

func prepareNestedTestFiles() afero.Fs {
	fs := prepareSingleTestFile()
	writeTestFile(fs, childFilePath, childFileOriginal)
	writeTestFile(fs, grandchildFilePath, grandchildFileOriginal)
	return fs
}

const (
	singleFilePath = "main.go"

	singleFileOriginal = `
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

	childFilePath = "subpackage/foo.go"

	childFileOriginal = `
package main

import (
	"github.com/some/package"
	"foobar.com/useful/package"
	"acme.com/awesome/package"
)
`
	childFileAcme = `
package main

import (
	"acme.com/awesome/package"

	"foobar.com/useful/package"
	"github.com/some/package"
)
`
	grandchildFilePath = "subpackage/nested/file.go"

	grandchildFileOriginal = `
package main

import (
	"github.com/foreign/package"
	"party.com/important/package"
	"acme.com/awesome/package"
	"time"
)
`
	grandchildFileAcme = `
package main

import (
	"time"

	"acme.com/awesome/package"

	"github.com/foreign/package"
	"party.com/important/package"
)
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

var (
	callerRegexp   = regexp.MustCompile(`.+/(.+?)\(.+\)`)
	callSiteRegexp = regexp.MustCompile(`\s+(.+?:\d+) `)
)
