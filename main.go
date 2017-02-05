package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

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
