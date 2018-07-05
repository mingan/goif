# Go import formatter - goif

Goif groups and orders imports in your Go source files. By default, it separates imports into two groups: standard library and everything else, each is then sorted. If you provide `--prefix` option, it will make a third group between these two containing all imports starting with this prefix.

## Example

```go
package main

import (
	"github.com/pressly/sup"
	"github.com/some/package"

	"github.com/different/library"

	"fmt"
	"github.com/some/otherpackage"
)
```

is with `--prefix=github.com/some` transformed into

```go
package main

import (
	"fmt"

	"github.com/some/otherpackage"
	"github.com/some/package"

	"github.com/different/library"
	"github.com/pressly/sup"
)
```

## Installation

`go get -u github.com/mingan/goif/cmd/goif`

## Usage

By default, `vendor/` folder is ignored by default. List of paths to exclude can be provided as `--exclude=vendor,something,some/nested/prefix`, the prefixes are evaluated relative to the current working directory.

I recommend setting goif to be run as a [Git pre-commit hook](https://git-scm.com/book/gr/v2/Customizing-Git-Git-Hooks), after `go fmt`.

## Contribution

Contributions and issues are welcome.