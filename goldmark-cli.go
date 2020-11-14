package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/yuin/goldmark"
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
	fmt.Fprintf(out, "Usage: gm [options] [file.md].\n")
	fmt.Fprintf(out, "  If the markdown file is missing the standard input is used in place.\n")
	fmt.Fprintf(out, "  The available options are:\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(out, "\n")
}

var (
	// The flags (for descriptions check SetParameters function)
	infile    string
	css       string
	title     string
	htmlshell string
	showhelp  bool

	// temp variable for error catch
	err error
)

func SetParameters() {
	flag.StringVarP(&css, "css", "s", "github", "The css file or the theme name present in github.com/kpym/markdown-css")
	flag.StringVarP(&title, "title", "t", "", "The page title.")
	flag.StringVar(&htmlshell, "html", "", "The html htmlshell (file or string).")
	flag.BoolVarP(&showhelp, "help", "h", false, "Print this help message.")
	// keep the flags order
	flag.CommandLine.SortFlags = false
	// in case of error do not display second time
	flag.CommandLine.Init("marianne", flag.ContinueOnError)
	// The help message
	flag.Usage = help
	err = flag.CommandLine.Parse(os.Args[1:])
	// affiche l'aide si demandé ou si erreur de paramètre
	if showhelp || err != nil {
		flag.Usage()
		check(err, "Problem parsing parameters.")
		os.Exit(0)
	}

	// chack for positional parameters
	if flag.NArg() > 1 {
		check(errors.New("No more than one positional parameter (markdown filename) can be specified."))
	}
	// get the positional parameter if any
	if flag.NArg() > 0 {
		infile = flag.Arg(0)
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
	err = goldmark.Convert(markdown, &mdhtml)
	check(err, "Problem parsing your markdown to html with goldmark.")
	// combine the template and the resulting
	var data = make(map[string]template.HTML)
	data["title"] = template.HTML(title)
	data["css"] = template.HTML(css)
	data["html"] = template.HTML(mdhtml.String())
	var finalhtml bytes.Buffer
	err = t.Execute(&finalhtml, data)
	check(err, "Problem building the HTML from the template and your markdown.")

	// output the result
	os.Stdout.Write(finalhtml.Bytes())
}
