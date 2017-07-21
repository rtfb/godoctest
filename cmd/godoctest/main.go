package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/rtfb/godoctest"
)

func runGoTest(workdir string) error {
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
		}
		return err
	}
	return nil
}

func main() {
	e := godoctest.NewExtractor()
	fcs := e.Run("testdata")
	testCode, err := godoctest.GenPkgTests(fcs[0])
	if err != nil {
		log.Fatalf("Failed to generate test code: %s\n", err)
	}
	err = ioutil.WriteFile(fcs[0].TestFileName(), testCode, 0666)
	if err != nil {
		log.Fatalf("Failed to write the generated tests file: %s\n", err)
	}
	err = runGoTest("testdata")
	if err != nil {
		log.Fatalf("Test execution failed: %s\n", err)
	}
}
