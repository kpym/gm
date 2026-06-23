package main

import (
	"fmt"
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
			outfile = infile[:len(infile)-9] + "index.html"
		} else {
			// otherwise we just change the extension
			outfile = infile[:len(infile)-3] + ".html"
		}
		if hasoutdir && filepath.IsLocal(outfile) {
			outfile = filepath.Join(outdir, outfile)
		}
		if err := os.MkdirAll(filepath.Dir(outfile), os.ModePerm); err != nil {
			check(err, "Problem to reach/create folder:", filepath.Dir(outfile))
		}
		err = os.WriteFile(outfile, html, 0644)
		check(err, "Problem modifying", outfile)
	}
}

// pathHasDot checks if the given path has a part starting with a dot (.)
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

// relPath returns the relative path from dir to path
// it returns an error if path is not inside dir,
// or if there is a problem getting the absolute paths
func relPath(dir, path string) (string, error) {
	var err error
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return "", err
	}
	return rel, nil
}

// pathInside checks if the given path or pattern is inside the given directory
func pathInside(dir, path string) bool {
	rel, err := relPath(dir, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
}

// isFile checks if the given path is a file (not a directory, not a pattern)
func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// isFolder checks if the given path is a folder (not a file, not a pattern)
func isFolder(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// buildFiles convert all .md files verifying one of the patterns to .html
func buildFiles() {
	var err error
	// get the current directory as a filesystem, needed for doublestar.Glob
	dirFS := os.DirFS(cwd)
	// decide if we are moving files or just building the .md files to .html
	movefiles := move && hasoutdir
	// check all patterns
	action := "Building"
	if movefiles {
		action = "Building and moving"
	}
	info(action+" files from '%s' to '%s'.\n", cwd, outdir)
	for _, pattern := range inpatterns {
		// skip folders, we only want files
		if isFolder(pattern) {
			continue
		}

		// if the input is piped
		if pattern == "stdin" {
			info("  Building from stdin...\n")
			buildMd("")
			continue
		}

		// the files corresponding to the pattern
		var allfiles []string
		if isFile(pattern) {
			// the .md files are built if they are inside the current directory
			// or if there is no output folder specified
			if !pathInside(cwd, pattern) {
				if !strings.HasSuffix(pattern, ".md") {
					if movefiles {
						try(fmt.Errorf("Input file '%s' is not inside the current directory '%s' and is not a markdown file.", pattern, cwd))
					}
					continue
				}
				if hasoutdir {
					try(fmt.Errorf("Input file '%s' is not inside the current directory '%s' and output folder is specified.", pattern, cwd))
					continue
				}
			}
			// pattern can be absolute outside the current directory
			// or relative inside the current directory
			info("  Treating %s...\n", pattern)
			allfiles = []string{pattern}
		} else {
			info("Looking for pattern '%s'.\n", pattern)
			// look for all files with the given patterns
			allfiles, err = doublestar.Glob(dirFS, pattern, doublestar.WithFilesOnly(), doublestar.WithNoFollow())
			check(err, "Problem looking for file pattern:", pattern)
			if len(allfiles) == 0 {
				info("No files found.\n")
				continue
			}
		}
		for _, infile := range allfiles {
			infile = filepath.Clean(infile)
			if skipdot && pathHasDot(infile) {
				info("  Skipping dot file %s...\n", infile)
				continue
			}
			// Skip files that are in the output directory
			if hasoutdir && pathInside(outdir, infile) {
				info("  Skipping file %s inside output folder...\n", infile)
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
				if err = os.MkdirAll(filepath.Dir(outfile), os.ModePerm); err != nil {
					check(err, "Problem to reach/create folder:", filepath.Dir(outfile))
				}
				err = os.Rename(infile, outfile)
				check(err, "Problem moving", infile)
				info("done.\n")
			}
		}
	}
}
