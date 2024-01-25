package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// buildMd compiles the infile (xxx.md | stdin) to outfile (xxx.html | stdout)
func buildMd(infile string) {
	// get the dir for link replacement, if any
	dir := filepath.Dir(infile)
	// Get the input
	var input io.Reader
	if infile != "" {
		f, err := os.Open(infile)
		if err != nil {
			check(err, "Problem opening", infile)
			return
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
		dir = "."
	}
	// read the input
	markdown, err := io.ReadAll(input)
	check(err, "Problem reading the markdown.")

	//compile the input
	html, err := compile(markdown)
	check(err, "Problem compiling the markdown.")
	if localmdlinks {
		html = replaceLinks(html, dir)
	}

	// output the result
	if infile == "" {
		os.Stdout.Write(html)
	} else {
		outfile := filepath.Join(outdir, infile[:len(infile)-3]+".html")
		if os.MkdirAll(filepath.Dir(outfile), os.ModePerm) != nil {
			check(err, "Problem to reach/create folder:", filepath.Dir(outfile))
		}
		err = os.WriteFile(outfile, html, 0644)
		check(err, "Problem modifying", outfile)
	}
}

// buildFiles convert all .md files verifying one of the patterns to .html
func buildFiles() {
	// check all patterns
	for _, pattern := range inpatterns {
		info("Looking for '%s'.\n", pattern)
		// if the input is piped
		if pattern == "stdin" {
			buildMd("")
			continue
		}
		// get the current directory as a filesystem, needed for doublestar.Glob
		cwd, err := os.Getwd()
		check(err, "Problem getting the current directory.")
		dirFS := os.DirFS(cwd)
		// look for all files with the given patterns
		// but build only .md ones
		allfiles, err := doublestar.Glob(dirFS, pattern, doublestar.WithFilesOnly(), doublestar.WithNoFollow())
		check(err, "Problem looking for file pattern:", pattern)
		if len(allfiles) == 0 {
			info("No files found.\n")
			continue
		}
		for _, infile := range allfiles {
			if strings.HasSuffix(infile, ".md") {
				info("  Converting %s...", infile)
				buildMd(infile)
				info("done.\n")
			}
		}
	}
}
