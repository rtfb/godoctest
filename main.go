package main

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func makeFileName(fc *fileComments) string {
	fn := fc.fileName
	ext := path.Ext(fn)
	return fn[:len(fn)-len(ext)] + "_gdt_test.go"
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
	testCode := genPkgTests(fcs[0])
	err = ioutil.WriteFile(makeFileName(fcs[0]), []byte(testCode), 0666)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("go", "test")
	cmd.Dir = "testdata"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		switch typedErr := err.(type) {
		case *exec.ExitError:
			println(typedErr.String())
		default:
			panic(err)
		}
	}
}
