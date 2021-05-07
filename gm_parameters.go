package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
)

// Display the usage help message
func help() {
	// get the default error output
	var out = pflag.CommandLine.Output()
	// var out = os.Stderr
	// write the help message
	fmt.Fprintf(out, "gm (version: %s): a goldmark cli tool which is a thin wrapper around github.com/yuin/goldmark.\n\n", version)
	fmt.Fprintf(out, "Usage: gm [options] <file.md|file pattern|stdin>.\n\n")
	fmt.Fprintf(out, "  If not serving (no `--serve` or `-s` option is used):\n")
	fmt.Fprintf(out, "  - if  file pattern is used, only the mached .md files are used;\n")
	fmt.Fprintf(out, "  - the .md files are converted to .html with the same name;\n")
	fmt.Fprintf(out, "  - if the .html file exists it is overwritten.\n\n")
	fmt.Fprintf(out, "  The available options are:\n\n")
	pflag.PrintDefaults()
	fmt.Fprintf(out, "\n")
}

var (
	// serve flags
	serve     bool
	serveDir  string
	serveFile string

	// build flags
	outdir     string
	inpatterns []string

	// template flags
	css       string
	title     string
	htmlshell string

	mdTemplate *template.Template

	// the GoldMark flags
	attribute      bool
	definitionList bool
	footnote       bool
	linkify        bool
	strikethrough  bool
	table          bool
	taskList       bool
	typographer    bool
	unsafe         bool
	autoHeadingId  bool
	hardWraps      bool
	xhtml          bool
	// The following goldmark options are missing:
	// - Subscript / Superscript
	// - Ins
	// - Mark
	// - Inline footnote
	// - Compact style definition lists
	// - Emojis
	// - Abbreviations
	// - Code highlighting
	// - Math rendering

	mdParser goldmark.Markdown

	// post GoldMark flags
	localmdlinks bool

	// info flags
	quiet    bool
	showhelp bool
	info     func(string, ...interface{})
)

// SetParameters configure the global variables from the command line flags.
func SetParameters() {
	pflag.BoolVarP(&serve, "serve", "s", false, "Start serving local .md file(s). No html is saved.")
	pflag.Lookup("serve").NoOptDefVal = "true"

	pflag.StringVarP(&css, "css", "c", "github", "A css url or the theme name present in github.com/kpym/markdown-css")
	pflag.StringVarP(&title, "title", "t", "", "The default page title. Used if no h1 is found in the .md file.")
	pflag.StringVar(&htmlshell, "html", "", "The html template (file or string).")

	pflag.StringVarP(&outdir, "out-dir", "o", "", "The build output folder (created if not already existing, not used if `--serve`).")

	pflag.BoolVar(&attribute, "gm-attribute", true, "GoldMark option: allows to define attributes on some elements.")
	pflag.BoolVar(&autoHeadingId, "gm-auto-heading-id", true, "GoldMark option: enables auto heading ids.")
	pflag.BoolVar(&definitionList, "gm-definition-list", true, "GoldMark option: enables definition lists.")
	pflag.BoolVar(&footnote, "gm-footnote", true, "GoldMark option: enables footnotes.")
	pflag.BoolVar(&linkify, "gm-linkify", true, "GoldMark option: activates auto links.")
	pflag.BoolVar(&strikethrough, "gm-strikethrough", true, "GoldMark option: enables strike through.")
	pflag.BoolVar(&table, "gm-table", true, "GoldMark option: enables tables.")
	pflag.BoolVar(&taskList, "gm-task-list", true, "GoldMark option: enables task lists.")
	pflag.BoolVar(&typographer, "gm-typographer", true, "GoldMark option: activate punctuations substitution with typographic entities.")
	pflag.BoolVar(&unsafe, "gm-unsafe", true, "GoldMark option: enables raw html.")

	pflag.BoolVar(&hardWraps, "gm-hard-wraps", false, "GoldMark option: render newlines as <br>.")
	pflag.BoolVar(&xhtml, "gm-xhtml", false, "GoldMark option: render as XHTML.")

	pflag.BoolVar(&localmdlinks, "links-md2html", true, "Replace .md with .html in links to local files (not used if `--serve`).")

	pflag.BoolVarP(&quiet, "quiet", "q", false, "No errors, no info is printed. Return error code is still available.")
	pflag.BoolVarP(&showhelp, "help", "h", false, "Print this help message.")
	// keep the flags order
	pflag.CommandLine.SortFlags = false
	// in case of error do not display second time
	pflag.CommandLine.Init("goldmark-cli", pflag.ContinueOnError)
	// The help message
	pflag.Usage = help
	err := pflag.CommandLine.Parse(os.Args[1:])
	// display the help message if the flag is set or if there is an error
	if showhelp || err != nil {
		pflag.Usage()
		check(err, "Problem parsing parameters.")
		os.Exit(0)
	}

	// quiet or no
	if quiet {
		info = func(format string, a ...interface{}) {}
	} else {
		info = func(format string, a ...interface{}) { fmt.Fprintf(os.Stderr, format, a...) }
	}

	// set the css
	if css != "" && !strings.Contains(css, "/") && !strings.Contains(css, ".") {
		css = "https://kpym.github.io/markdown-css/" + css + ".min.css"
	}
	// set the template
	t, err := ioutil.ReadFile(htmlshell)
	if err == nil {
		htmlshell = string(t)
	}
	if htmlshell == "" {
		htmlshell = defaultHTMLTemplate
	}

	if serve {
		setServeParameters()
	} else {
		setBuildParameters()
	}

	setTemplate()
	setGoldMark()
}

// setServeParameters prepare the parameters to serve.
// if the positional parameter is like `path/file.md` then `path/` is served and `/file.md` is requested
// if the positional parameter is like `path/folder/` then `path/folder` is served and `/` is requested
func setServeParameters() {
	if pflag.NArg() > 1 {
		check(errors.New("Only one file or folder can be specified for serving."))
	}

	filename := "."
	if pflag.NArg() > 0 {
		filename = pflag.Arg(0)
	}
	fi, err := os.Stat(filename)
	check(err, "Can't access file or folder named", filename)
	switch mode := fi.Mode(); {
	case mode.IsDir():
		serveDir = filename
		serveFile = ""
	case mode.IsRegular():
		serveDir = filepath.Dir(filename)
		serveFile = filepath.Base(filename)
	default:
		check(fmt.Errorf("The specified path '%s'is not a file or folder.", filename))
	}

	// set the default title
	if title == "" {
		title = "GoldMark"
	}
}

// setBuildParameters get all paterns and create (if necessary) the "out dir".
func setBuildParameters() {
	// get the positional parameters
	inpatterns = pflag.Args()
	// check for positional parameters
	if len(inpatterns) == 0 {
		pflag.Usage()
		check(errors.New("At least one input 'file.md', 'p*ttern' or 'stdin' should be provided."))
	}

	// check the "out dir"
	if outdir != "" {
		if os.MkdirAll(outdir, os.ModePerm) != nil {
			check(fmt.Errorf("The specifide output folder '%s' is not reachable.", outdir))
		}
	}
}

// setGoldMark creates a new markdown parser with configuration based on the parameter flags.
// The code is borrowed from:
//		https://github.com/gohugoio/hugo/blob/d90e37e0c6e812f9913bf256c9c81aa05b7a08aa/markup/goldmark/convert.go
func setGoldMark() {
	var (
		rendererOptions []renderer.Option
		extensions      []goldmark.Extender
		parserOptions   []parser.Option
	)

	if attribute {
		parserOptions = append(parserOptions, parser.WithAttribute())
	}
	if autoHeadingId {
		parserOptions = append(parserOptions, parser.WithAutoHeadingID())
	}
	if definitionList {
		extensions = append(extensions, extension.DefinitionList)
	}
	if footnote {
		extensions = append(extensions, extension.Footnote)
	}
	if linkify {
		extensions = append(extensions, extension.Linkify)
	}
	if strikethrough {
		extensions = append(extensions, extension.Strikethrough)
	}
	if table {
		extensions = append(extensions, extension.Table)
	}
	if taskList {
		extensions = append(extensions, extension.TaskList)
	}
	if typographer {
		extensions = append(extensions, extension.Typographer)
	}
	if unsafe {
		rendererOptions = append(rendererOptions, html.WithUnsafe())
	}
	if hardWraps {
		rendererOptions = append(rendererOptions, html.WithHardWraps())
	}
	if xhtml {
		rendererOptions = append(rendererOptions, html.WithXHTML())
	}

	mdParser = goldmark.New(
		goldmark.WithExtensions(
			extensions...,
		),
		goldmark.WithParserOptions(
			parserOptions...,
		),
		goldmark.WithRendererOptions(
			rendererOptions...,
		),
	)
}

// setTemplate parse the `html` flag to template.
func setTemplate() {
	var err error
	mdTemplate, err = template.New("md").Parse(htmlshell)
	check(err, "Problem parsing the HTML template.")
}
