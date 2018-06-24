package goif

import (
	"sort"
	"io"
	"strings"
	"bufio"
	"regexp"
	"bytes"
	"errors"
	"fmt"
)

type formatter struct {
	orgGroupPrefix string
	importOpened   bool
	importLines    []importLine
}

func NewFormatter(orgGroupPrefix string) *formatter {
	return &formatter{
		orgGroupPrefix: orgGroupPrefix,
	}
}

func (f *formatter) Format(r io.Reader, w io.Writer) error {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err == nil || err == io.EOF {
			if err := f.line(line, w); err != nil {
				return err
			}
			if err == io.EOF {
				break
			}
		}
	}
	f.flushUnclosedBlock()
	return nil
}

func (f *formatter) line(line string, w io.Writer) error {
	switch {
	case !f.importOpened && isImportOpeningLine(line):
		f.importOpened = true
	case f.importOpened:
		if isImportClosingLine(line) {
			f.writeOrderedImports(w)
			f.importOpened = false
			f.importLines = []importLine{}
		} else if len(strings.TrimSpace(line)) > 0 {
			if id, err := parseImportLine(line); err == nil {
				f.importLines = append(f.importLines, id)
			} else {
				return err
			}
		}
	default:
		w.Write([]byte(line))
	}
	
	return nil
}

func isImportOpeningLine(line string) bool {
	return openingLineRegexp.MatchString(line)
}

func isImportClosingLine(line string) bool {
	return closingLineRegexp.MatchString(line)
}

func (f *formatter) writeOrderedImports(w io.Writer) {
	w.Write([]byte("import (\n"))
	groups := groupAndSort(f.orgGroupPrefix, f.importLines)

	count := len(groups)
	for i, group := range groups {
		for _, imp := range group {
			w.Write([]byte(imp.String()))
			w.Write([]byte("\n"))
		}
		if i < count-1 {
			w.Write([]byte("\n"))
		}
	}
	w.Write([]byte(")\n"))
}

func groupAndSort(orgPrefix string, imports []importLine) [][]importLine {
	var stdlib, org, others []importLine

	for _, imp := range imports {
		if strings.Contains(imp.importDecl.pckg, orgPrefix) {
			org = append(org, imp)
		} else if strings.Contains(imp.importDecl.pckg, ".") {
			others = append(others, imp)
		} else {
			stdlib = append(stdlib, imp)
		}
	}

	sort.Sort(sortableImports(org))
	sort.Sort(sortableImports(stdlib))
	sort.Sort(sortableImports(others))

	var out [][]importLine
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

func (f *formatter) flushUnclosedBlock() {
	if !f.importOpened {
		return
	}
}

type importLine struct {
	importDecl importDecl
	comment    comment
}

type importDecl struct {
	pckg  string
	alias string
}

func (id importDecl) empty() bool {
	return id.pckg == "" && id.alias == ""
}

type comment string

func parseImportLine(line string) (importLine, error) {
	if l, err := parseImportDecl(line); err == nil {
		return l, nil
	}
	return parseComment(line)
}

var (
	importDeclRegexp = regexp.MustCompile(`^\s*((\w+)\s+)?"(\S+)"`)
	commentRegexp    = regexp.MustCompile(`^\s*//(.+)`)
	openingLineRegexp = regexp.MustCompile(`^\s*import \(`)
	closingLineRegexp = regexp.MustCompile(`^\s*\)`)
)

func parseImportDecl(line string) (importLine, error) {
	matches := importDeclRegexp.FindStringSubmatch(line)
	if len(matches) == 0 {
		return importLine{}, errors.New("unrecognized import line")
	}

	return importLine{
		importDecl: importDecl{
			alias: matches[2],
			pckg:  matches[3],
		},
	}, nil
}

func parseComment(line string) (importLine, error) {
	matches := commentRegexp.FindStringSubmatch(line)
	if len(matches) == 0 {
		return importLine{}, errors.New("unrecognized import line")
	}

	return importLine{
		comment: comment(matches[1]),
	}, nil
}

func (il importLine) String() string {
	if !il.importDecl.empty() {
		buf := bytes.Buffer{}
		buf.WriteString("\t")
		if il.importDecl.alias != "" {
			buf.WriteString(il.importDecl.alias)
			buf.WriteString(" ")
		}
		buf.WriteString("\"")
		buf.WriteString(il.importDecl.pckg)
		buf.WriteString("\"")

		return buf.String()
	} else {
		return fmt.Sprintf("\t//%s", il.comment)
	}
}

type sortableImports []importLine

func (imps sortableImports) Len() int {
	return len(imps)
}

func (imps sortableImports) Less(i, j int) bool {
	return imps[i].importDecl.pckg < imps[j].importDecl.pckg
}

func (imps sortableImports) Swap(i, j int) {
	imps[i], imps[j] = imps[j], imps[i]
}
