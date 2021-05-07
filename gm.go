package main

import (
	"fmt"
	"os"
)

// version will be set by goreleaser based on the git tag
var version string = "--"

// printError acts only is error is present:
// - print the error message
// - panic if necessary
func printError(fatal bool, e error, m ...interface{}) {
	if e != nil {
		if len(m) > 0 {
			fmt.Fprint(os.Stderr, "Error: ")
			fmt.Fprintln(os.Stderr, m...)
		} else {
			fmt.Fprintln(os.Stderr, "Error.")
		}
		fmt.Fprintln(os.Stderr, e)
		if fatal {
			panic(e)
		}
	}
}

// check print the error message and panic if there is an error
func check(e error, m ...interface{}) {
	printError(true, e, m...)
}

// check print the error message (no panic) if there is an error
func try(e error, m ...interface{}) {
	printError(false, e, m...)
}

// end is the last function executed in this program.
func end() {
	// in case of error return status is 1
	if r := recover(); r != nil {
		os.Exit(1)
	}

	// the normal return status is 0
	os.Exit(0)
}

// main is the entry point
func main() {

	// error handling
	defer end()

	// check the flags and init parser
	SetParameters()

	// serve or build ?
	if serve {
		serveFiles()
	} else {
		buildFiles()
	}
}
