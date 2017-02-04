package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type commentData struct {
	pkg          *ast.Package
	fileName     string
	file         *ast.File
	comments     map[int]*ast.CommentGroup
	funcComments map[string]*ast.CommentGroup
}

func containsTestAnnotation(cg *ast.CommentGroup) bool {
	for _, c := range cg.List {
		if strings.Contains(c.Text, "@test") {
			return true
		}
	}
	return false
}

func extractComments(pkgs map[string]*ast.Package, fs *token.FileSet) []*commentData {
	var cds []*commentData
	for _, v := range pkgs {
		for fk, fv := range v.Files {
			var cd commentData
			cd.comments = map[int]*ast.CommentGroup{}
			cd.pkg = v
			cd.fileName = fk
			cd.file = fv
			for _, c := range fv.Comments {
				if !containsTestAnnotation(c) {
					continue
				}
				cd.comments[fs.Position(c.Pos()).Line] = c
			}
			cds = append(cds, &cd)
		}
	}
	return cds
}

func extractFuncs(cds []*commentData, fs *token.FileSet) {
	for _, cd := range cds {
		cd.funcComments = map[string]*ast.CommentGroup{}
		for _, d := range cd.file.Decls {
			declLine := fs.Position(d.Pos()).Line
			_, found := cd.comments[declLine+1]
			if !found {
				continue
			}
			switch typedDecl := d.(type) {
			case *ast.FuncDecl:
				cd.funcComments[typedDecl.Name.Name] = cd.comments[declLine+1]
			}
		}
		cd.comments = nil // release for GC
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
	cds := extractComments(pkgs, fs)
	extractFuncs(cds, fs)
	for _, cd := range cds {
		for k, v := range cd.funcComments {
			fmt.Printf("%s.%s:%d: %s\n%s\n", cd.pkg.Name, cd.fileName,
				fs.Position(v.Pos()).Line-1, k, cgToStr(v))
		}
	}
}
