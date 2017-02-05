package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strconv"
	"strings"
)

type fileComments struct {
	pkg          *ast.Package
	fileName     string
	file         *ast.File
	comments     map[int]*ast.CommentGroup
	funcComments map[string]funcData
}

type funcData struct {
	declLine    int
	commentLine int
	decl        *ast.FuncType
	comment     *ast.CommentGroup
}

func containsTestAnnotation(cg *ast.CommentGroup) bool {
	for _, c := range cg.List {
		if strings.Contains(c.Text, "@test") {
			return true
		}
	}
	return false
}

func extractComments(pkgs map[string]*ast.Package, fs *token.FileSet) []*fileComments {
	var fcs []*fileComments
	for _, v := range pkgs {
		for fk, fv := range v.Files {
			var fc fileComments
			fc.comments = map[int]*ast.CommentGroup{}
			fc.pkg = v
			fc.fileName = fk
			fc.file = fv
			for _, c := range fv.Comments {
				if !containsTestAnnotation(c) {
					continue
				}
				fc.comments[fs.Position(c.Pos()).Line] = c
			}
			fcs = append(fcs, &fc)
		}
	}
	return fcs
}

func extractFuncs(fcs []*fileComments, fs *token.FileSet) {
	for _, fc := range fcs {
		fc.funcComments = map[string]funcData{}
		for _, d := range fc.file.Decls {
			switch typedDecl := d.(type) {
			case *ast.FuncDecl:
				declLine := fs.Position(d.Pos()).Line
				endLine := fs.Position(typedDecl.Type.End()).Line
				_, found := fc.comments[endLine+1]
				if !found {
					continue
				}
				fc.funcComments[typedDecl.Name.Name] = funcData{
					declLine:    declLine,
					commentLine: endLine + 1,
					decl:        typedDecl.Type,
					comment:     fc.comments[endLine+1],
				}
			}
		}
		fc.comments = nil // release for GC
	}
}

func cgToStr(cg *ast.CommentGroup) string {
	var s bytes.Buffer
	for i, c := range cg.List {
		if i > 0 {
			s.WriteByte('\n')
		}
		s.WriteString(c.Text)
	}
	return s.String()
}

type templateFuncData struct {
	FuncName string
	Hash     string
}

func prepForTemplate(fcm map[string]funcData) []templateFuncData {
	var result []templateFuncData
	i := 0
	for k := range fcm {
		i++
		result = append(result, templateFuncData{
			FuncName: k,
			Hash:     strconv.Itoa(i),
		})
	}
	return result
}

func main() {
	fs := token.NewFileSet()
	// include tells parser.ParseDir which files to include.
	include := func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go")
	}
	pkgs, err := parser.ParseDir(fs, "testdata" /*pkg.Dir*/, include, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	fcs := extractComments(pkgs, fs)
	extractFuncs(fcs, fs)
	for _, fc := range fcs {
		for k, v := range fc.funcComments {
			fmt.Printf("%s.%s:%d: %s\n%s\n", fc.pkg.Name, fc.fileName,
				v.declLine, k, cgToStr(v.comment))
		}
	}
	println(genPkgTests(fcs[0]))
}
