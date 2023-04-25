package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// version will be set by goreleaser based on the git tag
var version string = "--"
var goldmarkVersion string = "--"

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
func mainEnd() {
	// in case of error return status is 1
	if r := recover(); r != nil {
		os.Exit(1)
	}

	// the normal return status is 0
	os.Exit(0)
}

// If we terminate with Ctrl/Cmd-C we call end()
func catchCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		info("Bye.\n")
		mainEnd()
	}()
}

// main is the entry point
func main() {

	// error handling
	defer mainEnd()
	// interrupt handling
	catchCtrlC()

	// check the flags and initialize the parser
	SetParameters()

	// serve or build ?
	if serve {
		serveFiles()
	} else {
		buildFiles()
	}
}
