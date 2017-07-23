package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/rtfb/godoctest"
)

var (
	runGoTestFlag   bool
	writeToFileFlag bool
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

func init() {
	tUsage := "run 'go test' after generating doctests"
	wUsage := "write results to file instead of stdout"
	flag.BoolVar(&runGoTestFlag, "t", false, tUsage)
	flag.BoolVar(&writeToFileFlag, "w", false, wUsage)
}

func run(path string) {
	e := godoctest.NewExtractor()
	fcs, err := e.Run(path)
	if err != nil {
		log.Fatalf("Failed to extract godoctests: %s\n", err)
	}
	testCode, err := godoctest.GenPkgTests(fcs[0])
	if err != nil {
		log.Fatalf("Failed to generate test code: %s\n", err)
	}
	if writeToFileFlag {
		err = ioutil.WriteFile(fcs[0].TestFileName(), testCode, 0666)
		if err != nil {
			log.Fatalf("Failed to write the generated tests file: %s\n", err)
		}
	} else {
		fmt.Println(string(testCode))
	}
	if runGoTestFlag {
		err = runGoTest(path)
		if err != nil {
			log.Fatalf("Test execution failed: %s\n", err)
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: godoctest [flags] [path...]\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	for _, path := range flag.Args() {
		run(path)
	}
}
