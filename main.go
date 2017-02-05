package main

import (
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
	println(genPkgTests(fcs[0]))
}
