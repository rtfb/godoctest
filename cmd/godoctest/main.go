package main

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/rtfb/godoctest"
)

func runGoTest(workdir string) {
	cmd := exec.Command("go", "test")
	cmd.Dir = workdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		switch typedErr := err.(type) {
		case *exec.ExitError:
			println(typedErr.String())
		default:
			panic(err)
		}
	}
}

func main() {
	e := godoctest.NewExtractor()
	fcs := e.Run("testdata")
	testCode := godoctest.GenPkgTests(fcs[0])
	err := ioutil.WriteFile(fcs[0].TestFileName(), []byte(testCode), 0666)
	if err != nil {
		panic(err)
	}
	runGoTest("testdata")
}
