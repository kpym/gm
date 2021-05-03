package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
)

// the version will be set by goreleaser based on the git tag
var version string = "--"

// temporary default template
// this should be moved out of this file
var defaultHTMLTemplate string = `
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    {{- with .css }}
    <link rel="stylesheet" type="text/css" href="{{.}}">
    {{- end }}
    {{- with .title }}
    <title>{{.}}</title>
    {{- end }}
  </head>
	<body>
		<article class="markdown-body">
{{.html}}
		</article>
  </body>
</html>
`

// Check for error
// - do nothing if no error
// - print the error message and panic if there is an error
func check(e error, m ...interface{}) {
	if e != nil {
		if len(m) > 0 {
			fmt.Print("Error: ")
			fmt.Println(m...)
		} else {
			fmt.Println("Error.")
		}
		fmt.Printf("More info: %v\n", e)
		panic(e)
	}
}

// This is the last function executed in this program.
func end() {
	// in case of error return status is 1
	if r := recover(); r != nil {
		os.Exit(1)
	}

	// the normal return status is 0
	os.Exit(0)
}

// Display the usage help message
func help() {
	// get the default error output
	var out = flag.CommandLine.Output()
	// var out = os.Stderr
	// write the help message
	fmt.Fprintf(out, "gm (version: %s): a goldmark cli tool which is a thin wrapper around github.com/yuin/goldmark.\n\n", version)
	fmt.Fprintf(out, "Usage: gm [options] <file.md|file pattern|stdin>.\n\n")
	fmt.Fprintf(out, "  If a file pattern is used, only the mached .md files are used.\n")
	fmt.Fprintf(out, "  The .md files are converted to .html with the same name.\n")
	fmt.Fprintf(out, "  If the .html file exists it is overwritten.\n")
	fmt.Fprintf(out, "  The available options are:\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(out, "\n")
}

var (
	// The flags (for descriptions check SetParameters function)
	inpattern string
	css       string
	title     string
	htmlshell string

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

	localmdlinks bool

	showhelp bool

	// temp variable for error catch
	err error
)

// Set the configuration variables from the command line flags
// The following options are missing
// - Subscript / Superscript
// - Ins
// - Mark
// - Inline footnote
// - Compact style definition lists
// - Emojis
// - Abbreviations
// - Code highlighting
// - Math rendering
func SetParameters() {
	flag.StringVarP(&css, "css", "s", "github", "The css file or the theme name present in github.com/kpym/markdown-css")
	flag.StringVarP(&title, "title", "t", "", "The page title.")
	flag.StringVar(&htmlshell, "html", "", "The html shell (file or string).")

	flag.BoolVar(&attribute, "attribute", true, "Allows to define attributes on some elements.")
	flag.BoolVar(&autoHeadingId, "auto-heading-id", true, "Enables auto heading ids.")
	flag.BoolVar(&definitionList, "definition-list", true, "Enables definition lists.")
	flag.BoolVar(&footnote, "footnote", true, "Enables footnotes.")
	flag.BoolVar(&linkify, "linkify", true, "Activates auto links.")
	flag.BoolVar(&strikethrough, "strikethrough", true, "Enables strike through.")
	flag.BoolVar(&table, "table", true, "Enables tables.")
	flag.BoolVar(&taskList, "task-list", true, "Enables task lists.")
	flag.BoolVar(&typographer, "typographer", true, "Activate punctuations substitution with typographic entities.")
	flag.BoolVar(&unsafe, "unsafe", true, "Enables raw html.")

	flag.BoolVar(&hardWraps, "hard-wraps", false, "Render newlines as <br>.")
	flag.BoolVar(&xhtml, "xhtml", false, "Render as XHTML.")

	flag.BoolVar(&localmdlinks, "links-md2html", true, "Convert links to local .md files to corresponding .html.")

	flag.BoolVarP(&showhelp, "help", "h", false, "Print this help message.")
	// keep the flags order
	flag.CommandLine.SortFlags = false
	// in case of error do not display second time
	flag.CommandLine.Init("goldmark-cli", flag.ContinueOnError)
	// The help message
	flag.Usage = help
	err = flag.CommandLine.Parse(os.Args[1:])
	// display the help message if the flag is set or if there is an error
	if showhelp || err != nil {
		flag.Usage()
		check(err, "Problem parsing parameters.")
		os.Exit(0)
	}

	// check for positional parameters
	if flag.NArg() > 1 {
		flag.Usage()
		check(errors.New("No more than one positional parameter (markdown filename or pattern) can be specified."))
	}
	if flag.NArg() < 1 {
		flag.Usage()
		check(errors.New("One positional parameter (markdown filename or pattern) should be provided."))
	}
	// get the positional parameter
	inpattern = flag.Arg(0)

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
}

// Create new markdown parser with configuration based on the parameter flags
// The code is borrowed from:
//		https://github.com/gohugoio/hugo/blob/d90e37e0c6e812f9913bf256c9c81aa05b7a08aa/markup/goldmark/convert.go
func newMarkdown() goldmark.Markdown {
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

	md := goldmark.New(
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

	return md
}

func build(infile string, t *template.Template) {
	// Get the input
	var input io.Reader
	if infile != "" {
		f, err := os.Open(infile)
		check(err, "Problem opening", infile)
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}
	// read the input
	markdown, err := ioutil.ReadAll(input)
	check(err, "Problem while reading the markdown.")
	// convert the markdown to html code
	var mdhtml bytes.Buffer
	md := newMarkdown()
	err = md.Convert(markdown, &mdhtml)
	check(err, "Problem parsing your markdown to html with goldmark.")
	// combine the template and the resulting
	var data = make(map[string]template.HTML)
	data["title"] = template.HTML(title)
	data["css"] = template.HTML(css)
	data["html"] = template.HTML(mdhtml.String())
	var finalhtml bytes.Buffer
	err = t.Execute(&finalhtml, data)
	check(err, "Problem building the HTML from the template and your markdown.")
	result := finalhtml.Bytes()
	// replace .md links with .html for local files
	if localmdlinks {
		link := regexp.MustCompile(`href\s*=\s*"([^"]+)\.md"`)
		result = link.ReplaceAllFunc(result, func(s []byte) []byte {
			filename := strings.Split(string(s), `"`)[1]
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				return s
			}
			return []byte(fmt.Sprintf(`href="%s.html"`, filename[:len(filename)-3]))
		})
	}
	// output the result
	if infile == "" {
		os.Stdout.Write(result)
	} else {
		outfile := infile[:len(infile)-3] + ".html"
		err = ioutil.WriteFile(outfile, result, 0644)
		check(err, "Problem modifying", outfile)
	}
}

// entry point & validation
func main() {
	// error handling
	defer end()
	// The flags
	SetParameters()
	// Prepare the template
	t, err := template.New("md").Parse(htmlshell)
	check(err, "Problem parsing the HTML template.")
	// if the input is piped
	if inpattern == "stdin" {
		build("", t)
		return
	}
	// look for all files with the given pattern
	// but build only .md ones
	allfiles, err := filepath.Glob(inpattern)
	check(err, "Problem looking for file pattern:", inpattern)
	if len(allfiles) == 0 {
		check(errors.New("No files found."), "Problem looking for file pattern:", inpattern)
	}
	for _, infile := range allfiles {
		if strings.HasSuffix(infile, ".md") {
			fmt.Printf("Converting %s...", infile)
			build(infile, t)
			fmt.Println("done.")
		}
	}
}
