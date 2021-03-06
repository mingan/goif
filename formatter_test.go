package goif

import (
	"testing"
	"os"
	"log"
	"bufio"
	"io"
	"bytes"
	
	"github.com/andreyvit/diff"
)

func TestFormatter(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		testFormatter(t, "", "", "")
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
			"enectiva.cz",
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

	t.Run("simple case with all and github prefix", func(t *testing.T) {
		testFormatter(
			t,
			`
package main

import (
	"github.com/pressly/sup"
	"github.com/some/package"

	"github.com/different/library"

	"fmt"
	"github.com/some/otherpackage"
)

func main() {
	fmt.Println("Hello world")
}
`,
			"github.com/some",
			`
package main

import (
	"fmt"

	"github.com/some/otherpackage"
	"github.com/some/package"

	"github.com/different/library"
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
			"enectiva.cz",
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
	_ "log"
	f "fmt"
)

func main() {
	fmt.Println("Hello world")
}
`,
			"enectiva.cz",
			`
package main

import (
	f "fmt"
	_ "log"
)

func main() {
	fmt.Println("Hello world")
}
`,
		)
	})

	t.Run("multiple import blocks", func(t *testing.T) {
		testFormatter(
			t,
			`
package main

import (
	"log"
	"fmt"
)

import (
	"time"
	"math"
)

func main() {
	fmt.Println("Hello world")
}
`,
			"enectiva.cz",
			`
package main

import (
	"fmt"
	"log"
)

import (
	"math"
	"time"
)

func main() {
	fmt.Println("Hello world")
}
`,
		)
	})

	t.Run("no empty lines", func(t *testing.T) {
		testFormatter(
			t,
			`package main
import (
	"log"
	"fmt"
)`,
			"enectiva.cz",
			`package main
import (
	"fmt"
	"log"
)
`,
		)
	})

	t.Run("lone import is ignored", func(t *testing.T) {
		testFormatter(
			t,
			`
package main

import "log"
`,
			"enectiva.cz",
			`
package main

import "log"
`,
		)
	})

	t.Run("harmless comments in the middle are grouped", func(t *testing.T) {
		testFormatter(
			t,
			`
package main

import (
// comment 1
	"log"
	// comment 2
	"fmt"
	// comment 3
)
`,
			"enectiva.cz",
			`
package main

import (
	// comment 1
	// comment 2
	// comment 3
	"fmt"
	"log"
)
`,
		)
	})

	t.Run("commented out import block", func(t *testing.T) {
		testFormatter(
			t,
			`package main

// import (
//	"log"
//	"fmt"
// )
`,
			"enectiva.cz",
			`package main

// import (
//	"log"
//	"fmt"
// )
`,
		)
	})

	t.Run("file comment", func(t *testing.T) {
		testFormatter(
			t,
			`// +build linux

package main

import (
	"log"
	"fmt"
)
`,
			"enectiva.cz",
			`// +build linux

package main

import (
	"fmt"
	"log"
)
`,
		)
	})

	t.Run("unclosed import block", func(t *testing.T) {

		t.Parallel()

		formatter := NewFormatter("enectiva.cz")

		var output bytes.Buffer
		err := formatter.Format(bytes.NewBufferString(`package main
import (
	"log"
	"fmt"

func main() {
	fmt.Println("I won't compile")
}
`), &output)
		if err == nil {
			t.Error("Expected to fail")
		}
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
			p.line(line, target)
			if err == io.EOF {
				break
			}
		}
	}
}

func testFormatter(t *testing.T, input, prefix, expected string) {
	t.Parallel()

	formatter := NewFormatter(prefix)

	var output bytes.Buffer
	formatter.Format(bytes.NewBufferString(input), &output)

	if output.String() != expected {
		t.Error(diff.LineDiff(expected, output.String()))
	}
}
