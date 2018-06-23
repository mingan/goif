package main

import (
	"fmt"
	"testing"
	"os"
	"log"
	"bufio"
	"io"
	"bytes"
	"github.com/andreyvit/diff"
)

// multiple import blocks
// import alias
// single line
// white space
// unclosed block
// commented out blocks

func TestFormatter(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		testFormatter(t, "", "")
	})

	t.Run("simple case with all groups", func(t *testing.T) {
		testFormatter(
			t,
			`
package main

import (
	"github.com/pressly/sup"
	"enectiva.cz/prefab/log"

	"enectiva.cz/prefab/api"

	"fmt"
	"enectiva.cz/prefab/advock"
)

func main() {
	fmt.Println("Hello world")
}
`,
			`
package main

import (
	"fmt"

	"enectiva.cz/prefab/advock"
	"enectiva.cz/prefab/api"
	"enectiva.cz/prefab/log"

	"github.com/pressly/sup"
)

func main() {
	fmt.Println("Hello world")
}
`,
		)
	})
	
	t.Run("only one group", func(t *testing.T) {
		testFormatter(
			t,
			`
package main

import (
	"log"
	"fmt"
)

func main() {
	fmt.Println("Hello world")
}
`,
			`
package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Hello world")
}
`,
		)
	})

	t.Run("import alias", func(t *testing.T) {
		testFormatter(
			t,
			`
package main

import (
	"log"
	f "fmt"
)

func main() {
	fmt.Println("Hello world")
}
`,
			`
package main

import (
	f "fmt"
	"log"
)

func main() {
	fmt.Println("Hello world")
}
`,
		)
	})

	return

	source, err := os.OpenFile("sample/sample.go", os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	stat, err := source.Stat()
	if err != nil {
		log.Fatal(err)
	}

	target, err := os.Create("sample/sample_formatted.go")
	if err != nil {
		log.Fatal(err)
	}
	defer target.Close()
	target.Chmod(stat.Mode().Perm())

	p := formatter{}

	reader := bufio.NewReader(source)
	for {
		line, err := reader.ReadString('\n')
		if err == nil || err == io.EOF {
			p.Line(line, target)
			if err == io.EOF {
				fmt.Println("done reading", line)
				break
			}
		}

		//		fmt.Println(strings.TrimRight(line, "\n"), p)
	}
}
func testFormatter(t *testing.T, input, expected string) {
	t.Parallel()

	formatter := NewFormatter("enectiva.cz")

	var output bytes.Buffer
	formatter.Format(bytes.NewBufferString(input), &output)

	if output.String() != expected {
		t.Error(diff.LineDiff(expected, output.String()))
	}
}
