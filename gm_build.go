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

	// Read the input
	markdown, err := io.ReadAll(input)
	check(err, "Problem reading the markdown.")

	// Apply re-md rules if available
	if len(reMdRules) > 0 {
		markdown = reMdRules.Apply(markdown)
	}

	// compile the input
	html, err := compile(markdown)
	check(err, "Problem compiling the markdown.")
	if localmdlinks {
		html = replaceLinks(html, dir)
	}

	// Apply re-html rules if available
	if len(reHtmlRules) > 0 {
		html = reHtmlRules.Apply(html)
	}

	// output the result
	if infile == "" {
		os.Stdout.Write(html)
	} else {
		var outfile string

		if readme && strings.ToLower(filepath.Base(infile)) == "readme.md" {
			// if it is a README.md file, we want to name it index.html
			outfile = filepath.Join(outdir, infile[:len(infile)-9]+"index.html")
		} else {
			// otherwise we just change the extension
			outfile = filepath.Join(outdir, infile[:len(infile)-3]+".html")
		}
		if os.MkdirAll(filepath.Dir(outfile), os.ModePerm) != nil {
			check(err, "Problem to reach/create folder:", filepath.Dir(outfile))
		}
		err = os.WriteFile(outfile, html, 0644)
		check(err, "Problem modifying", outfile)
	}
}

func pathFirstPart(path string) string {
	i := 0
	for ; i < len(path); i++ {
		if os.IsPathSeparator(path[i]) {
			break
		}
	}
	if i == len(path) {
		return path + string(os.PathSeparator)
	}
	return path[:i+1]
}

func pathHasDot(path string) bool {
	wasSeparator := true
	for i := 0; i < len(path); i++ {
		if path[i] == '.' && wasSeparator {
			return true
		}
		wasSeparator = os.IsPathSeparator(path[i])
	}
	return false
}

// buildFiles convert all .md files verifying one of the patterns to .html
func buildFiles() {
	// get the current directory
	cwd, err := os.Getwd()
	check(err, "Problem getting the current directory.")
	// get the current directory as a filesystem, needed for doublestar.Glob
	dirFS := os.DirFS(cwd)
	// normalize the output directory and set movefiles and outstart
	outdir, err = filepath.Abs(outdir)
	check(err, "Problem getting the absolute path of the output directory.")
	movefiles := move && outdir != cwd
	outdir, err = filepath.Rel(cwd, outdir)
	check(err, "Problem getting the relative path of the output directory.")
	// get the first part of the relative out path
	outstart := pathFirstPart(outdir)
	// check all patterns
	action := "Building"
	if movefiles {
		action = "Building and moving"
	}
	info(action+" files from '%s' to '%s'.\n", cwd, outdir)
	for _, pattern := range inpatterns {
		info("Looking for '%s'.\n", pattern)
		// if the input is piped
		if pattern == "stdin" {
			buildMd("")
			continue
		}
		// look for all files with the given patterns
		// but build only .md ones
		allfiles, err := doublestar.Glob(dirFS, pattern, doublestar.WithFilesOnly(), doublestar.WithNoFollow())
		check(err, "Problem looking for file pattern:", pattern)
		if len(allfiles) == 0 {
			info("No files found.\n")
			continue
		}
		for _, infile := range allfiles {
			infile = filepath.Clean(infile)
			if skipdot && pathHasDot(infile) {
				info("  Skipping %s...\n", infile)
				continue
			}
			if strings.HasPrefix(infile, outstart) {
				continue
			}
			if strings.HasSuffix(infile, ".md") {
				info("  Converting %s...", infile)
				buildMd(infile)
				info("done.\n")
			} else if movefiles {
				// move the file if it is not markdown and not already in the output folder
				info("  Moving %s...", infile)
				outfile := filepath.Join(outdir, infile)
				if os.MkdirAll(filepath.Dir(outfile), os.ModePerm) != nil {
					check(err, "Problem to reach/create folder:", filepath.Dir(outfile))
				}
				err := os.Rename(infile, outfile)
				check(err, "Problem moving", infile)
				info("done.\n")
			}
		}
	}
}
