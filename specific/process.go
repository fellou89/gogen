// Package specific copies the source from a package and generates a second
// package replacing some of the types used. It's aimed at taking generic
// packages that rely on interface{} and generating packages that use a
// specific type.
package specific

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

// Options struct
type Options struct {
	SkipTestFiles bool
}

// DefaultOptions struct
var DefaultOptions = Options{
	SkipTestFiles: false,
}

// Process creates a specific package from the generic specified in pkg
func Process(pkg, outdir, verb, newType string, optset ...func(*Options)) error {
	opts := DefaultOptions
	for _, fn := range optset {
		fn(&opts)
	}

	p, err := findPackage(pkg)
	if err != nil {
		return err
	}

	if outdir == "" {
		outdir = path.Base(pkg)
	}

	if verb == "" {
		return errors.New("Need a REST verb")
	}
	verb = strings.ToLower(verb)

	if err := os.MkdirAll(outdir, os.ModePerm); err != nil {
		return err
	}

	t := parseTargetType(newType)

	files, err := processFiles(p, verb, p.GoFiles, t)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	if err := write(t, verb, outdir, files); err != nil {
		return err
	}

	if opts.SkipTestFiles {
		return nil
	}

	files, err = processFiles(p, verb, p.TestGoFiles, t)
	if err != nil {
		return err
	}

	return write(t, verb, outdir, files)
}

func isPlural(t targetType) bool {
	return t.newType[len(t.newType)-1:] == "s"
}

func processFiles(p Package, verb string, files []string, t targetType) ([]processedFile, error) {
	var result []processedFile
	for _, f := range files {
		res, err := processFile(p, f, t, verb)
		if err != nil {
			return result, err
		}
		result = append(result, res)
	}
	return result, nil
}

func processFile(p Package, filename string, t targetType, v string) (processedFile, error) {
	res := processedFile{filename: filename}

	in, err := os.Open(path.Join(p.Dir, filename))
	if err != nil {
		return res, FileError{Package: p.Dir, File: filename, Err: err}
	}
	src, err := ioutil.ReadAll(in)
	if err != nil {
		return res, FileError{Package: p.Dir, File: filename, Err: err}
	}

	res.fset = token.NewFileSet()
	res.file, err = parser.ParseFile(res.fset, res.filename, src, parser.ParseComments|parser.AllErrors|parser.DeclarationErrors)
	if err != nil {
		return res, FileError{Package: p.Dir, File: filename, Err: err}
	}

	if replace(t, res.file, v) && t.newPkg != "" {
		astutil.AddImport(res.fset, res.file, t.newPkg)
	}

	return res, err
}

func replace(t targetType, n ast.Node, verb string) (replaced bool) {
	verb = strings.ToUpper(verb[:1]) + verb[1:]

	ast.Walk(visitFn(func(node ast.Node) {
		if node == nil {
			return
		}

		switch n := node.(type) {
		case *ast.Comment:
			n.Text = strings.Replace(n.Text, "Temp", t.newType, 1)
			n.Text = strings.Replace(n.Text, "Verb", verb, 1)

		case *ast.Ident:
			n.Name = strings.Replace(n.Name, "templates", "service", 1)

			n.Name = strings.Replace(n.Name, "Temp", t.newType, 1)
			n.Name = strings.Replace(n.Name, strings.ToLower("Temp"), strings.ToLower(t.newType), 1)

			n.Name = strings.Replace(n.Name, "Verb", verb, 1)
		}
	}), n)
	return replaced
}

type visitFn func(node ast.Node)

func (fn visitFn) Visit(node ast.Node) ast.Visitor {
	fn(node)
	return fn
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func write(t targetType, verb, outdir string, files []processedFile) error {
	verb = strings.ToUpper(verb[:1]) + verb[1:]

	for _, f := range files {
		out, err := os.Create(path.Join(outdir, strings.Replace(f.filename, "temp", strings.ToLower(verb)+"_"+toSnakeCase(t.newType), 1)))

		if err != nil {
			return FileError{Package: outdir, File: f.filename, Err: err}
		}

		fmt.Fprintf(out, "// +build "+verb+t.newType+"\n\n")
		printer.Fprint(out, f.fset, f.file)
	}
	return nil
}

// FileError Struct
type FileError struct {
	Package string
	File    string
	Err     error
}

func (ferr FileError) Error() string {
	return fmt.Sprintf("error in %s: %s", path.Join(ferr.Package, ferr.File), ferr.Err.Error())
}

type processedFile struct {
	filename string
	fset     *token.FileSet
	file     *ast.File
}
