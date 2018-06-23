package main

import (
	"sort"
	"io"
	"strings"
	"bufio"
)

type formatter struct {
	orgGroupPrefix string
	importOpened   bool
	importLines    []string
}

func NewFormatter(orgGroupPrefix string) *formatter {
	return &formatter{
		orgGroupPrefix: orgGroupPrefix,
	}
}

func (f *formatter) Format(r io.Reader, w io.Writer) {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err == nil || err == io.EOF {
			f.Line(line, w)
			if err == io.EOF {
				break
			}
		}
	}
}

func (f *formatter) Line(line string, w io.Writer) {
	switch {
	case !f.importOpened && importOpeningLine(line):
		f.importOpened = true
	case f.importOpened:
		if importClosingLine(line) {
			f.writeOrderedImports(w)
			f.importOpened = false
		} else if len(strings.TrimSpace(line)) > 0 {
			f.importLines = append(f.importLines, line)
		}
	default:
		w.Write([]byte(line))
	}
}

func importOpeningLine(line string) bool {
	return strings.Contains(line, "import (")
}

func importClosingLine(line string) bool {
	return strings.Contains(line, ")")
}

func (f *formatter) writeOrderedImports(w io.Writer) {
	w.Write([]byte("import (\n"))
	groups := groupAndSort(f.orgGroupPrefix, f.importLines)

	count := len(groups)
	for i, group := range groups {
		for _, imp := range group {
			w.Write([]byte(imp))
		}
		if i < count-1 {
			w.Write([]byte("\n"))
		}
	}
	w.Write([]byte(")\n"))
}

func groupAndSort(orgPrefix string, imports []string) [][]string {
	var stdlib, org, others []string

	for _, imp := range imports {
		if strings.Contains(imp, orgPrefix) {
			org = append(org, imp)
		} else if strings.Contains(imp, ".") {
			others = append(others, imp)
		} else {
			stdlib = append(stdlib, imp)
		}
	}

	sort.Sort(sortableImports(org))
	sort.Sort(sortableImports(stdlib))
	sort.Sort(sortableImports(others))

	var out [][]string
	if len(stdlib) > 0 {
		out = append(out, stdlib)
	}
	if len(org) > 0 {
		out = append(out, org)
	}
	if len(others) > 0 {
		out = append(out, others)
	}

	return out
}

type sortableImports []string

func (imps sortableImports) Len() int {
	return len(imps)
}

func (imps sortableImports) Less(i, j int) bool {
	return imps[i] < imps[j]
}

func (imps sortableImports) Swap(i, j int) {
	imps[i], imps[j] = imps[j], imps[i]
}
