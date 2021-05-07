package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/kpym/goldmark-cli/internal/browser"
	"github.com/spf13/pflag"
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
var defaultHTMLTemplate string = `<!DOCTYPE html>
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

//go:embed md.png
var favIcon []byte

// Check for error
// - do nothing if no error
// - print the error message and panic if there is an error
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

// Check for error
// - do nothing if no error
// - print the error message and panic if there is an error
func check(e error, m ...interface{}) {
	printError(true, e, m...)
}

// try will log the error message if any
func try(e error, m ...interface{}) {
	printError(false, e, m...)
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
	var out = pflag.CommandLine.Output()
	// var out = os.Stderr
	// write the help message
	fmt.Fprintf(out, "gm (version: %s): a goldmark cli tool which is a thin wrapper around github.com/yuin/goldmark.\n\n", version)
	fmt.Fprintf(out, "Usage: gm [options] <file.md|file pattern|stdin>.\n\n")
	fmt.Fprintf(out, "  If a file pattern is used, only the mached .md files are used.\n")
	fmt.Fprintf(out, "  The .md files are converted to .html with the same name.\n")
	fmt.Fprintf(out, "  If the .html file exists it is overwritten.\n")
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

	mdParser goldmark.Markdown

	// post GoldMark flags
	localmdlinks bool

	// info flags
	quiet    bool
	showhelp bool
	info     func(string, ...interface{})
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
	pflag.BoolVarP(&serve, "serve", "s", false, "Start serving local .md file(s). No html is saved.")
	pflag.Lookup("serve").NoOptDefVal = "true"

	pflag.StringVarP(&css, "css", "c", "github", "A css url or the theme name present in github.com/kpym/markdown-css")
	pflag.StringVarP(&title, "title", "t", "", "The page title.")
	pflag.StringVar(&htmlshell, "html", "", "The html template (file or string).")

	pflag.StringVarP(&outdir, "out-dir", "o", "", "The build output folder (created if not already existing).")

	pflag.BoolVar(&attribute, "attribute", true, "Allows to define attributes on some elements.")
	pflag.BoolVar(&autoHeadingId, "auto-heading-id", true, "Enables auto heading ids.")
	pflag.BoolVar(&definitionList, "definition-list", true, "Enables definition lists.")
	pflag.BoolVar(&footnote, "footnote", true, "Enables footnotes.")
	pflag.BoolVar(&linkify, "linkify", true, "Activates auto links.")
	pflag.BoolVar(&strikethrough, "strikethrough", true, "Enables strike through.")
	pflag.BoolVar(&table, "table", true, "Enables tables.")
	pflag.BoolVar(&taskList, "task-list", true, "Enables task lists.")
	pflag.BoolVar(&typographer, "typographer", true, "Activate punctuations substitution with typographic entities.")
	pflag.BoolVar(&unsafe, "unsafe", true, "Enables raw html.")

	pflag.BoolVar(&hardWraps, "hard-wraps", false, "Render newlines as <br>.")
	pflag.BoolVar(&xhtml, "xhtml", false, "Render as XHTML.")

	pflag.BoolVar(&localmdlinks, "links-md2html", true, "Convert links to local .md files to the corresponding .html.")

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
		info("Will serve folder:%s\n", serveDir)
	case mode.IsRegular():
		serveDir = filepath.Dir(filename)
		serveFile = filepath.Base(filename)
		info("Will serve file %s from folder:%s\n", serveFile, serveDir)
	default:
		check(fmt.Errorf("The specified path '%s'is not a file or folder.", filename))
	}

	// set the default title
	if title == "" {
		title = "GoldMark"
	}
}

func setBuildParameters() {
	// get the positional parameters
	inpatterns = pflag.Args()
	// check for positional parameters
	if len(inpatterns) == 0 {
		pflag.Usage()
		check(errors.New("At least one input 'file.md', 'p*ttern' or 'stdin' should be provided."))
	}

	// check the out dir
	if outdir != "" {
		if os.MkdirAll(outdir, os.ModePerm) != nil {
			check(fmt.Errorf("The specifide output folder '%s' is not reachable.", outdir))
		}
	}
}

// Create new markdown parser with configuration based on the parameter flags
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

func setTemplate() {
	var err error
	mdTemplate, err = template.New("md").Parse(htmlshell)
	check(err, "Problem parsing the HTML template.")
}

var regexTitle = regexp.MustCompile(`(?m)^\#\s(.+)$`)

// getTitle search for the first h1 title in the markdown
// if it can't find it returns the default one
func getTitle(markdown []byte) string {
	res := regexTitle.FindSubmatch(markdown)
	if len(res) > 1 {
		return string(res[1])
	}

	return title
}

// compile convert markdown to full html
// by first applying markdown
// and then integrating the result in a html template
func compile(markdown []byte) (html []byte, err error) {
	// temporary buffer
	var htmlBuf bytes.Buffer

	// convert md to html code
	err = mdParser.Convert(markdown, &htmlBuf)
	if err != nil {
		return nil, fmt.Errorf("Problem parsing markdown code to html with goldmark.\n %w", err)
	}

	// combine the template and the resulting
	var data = make(map[string]template.HTML)
	data["title"] = template.HTML(getTitle(markdown))
	data["css"] = template.HTML(css)
	data["html"] = template.HTML(htmlBuf.String())

	htmlBuf.Reset()
	err = mdTemplate.Execute(&htmlBuf, data)
	if err != nil {
		return nil, fmt.Errorf("Problem building HTML from template.\n %w", err)
	}

	return htmlBuf.Bytes(), nil
}

var regexMdLink = regexp.MustCompile(`href\s*=\s*"([^"]+)\.md"`)

func replaceLinks(html []byte, dir string) []byte {
	// replace .md links with .html for local files
	return regexMdLink.ReplaceAllFunc(html, func(s []byte) []byte {
		filename := strings.Split(string(s), `"`)[1]
		relname := filepath.Join(dir, filename)
		if _, err := os.Stat(relname); err != nil {
			return s
		}
		return []byte(fmt.Sprintf(`href="%s.html"`, filename[:len(filename)-3]))
	})
}

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
	markdown, err := ioutil.ReadAll(input)
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
		err = ioutil.WriteFile(outfile, html, 0644)
		check(err, "Problem modifying", outfile)
	}
}

// buildFiles all .md files verifying one of the patterns to .html
func buildFiles() {
	// check all patterns
	for _, pattern := range inpatterns {
		info("Looking for '%s'.\n", pattern)
		// if the input is piped
		if pattern == "stdin" {
			buildMd("")
			continue
		}
		// look for all files with the given patterns
		// but build only .md ones
		allfiles, err := filepath.Glob(pattern)
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

// availablePort provides the first available port after 8080
// or 8180 if no available ports are present.
func availablePort() (port string) {
	for i := 8080; i < 8181; i++ {
		port = strconv.Itoa(i)
		if ln, err := net.Listen("tcp", "localhost:"+port); err == nil {
			ln.Close()
			break
		}
	}
	return port
}

// serveFiles the local folder
// all requests to .md or .html file check if the .md file exists
// if yes it is compiled and send as html
func serveFiles() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filename := filepath.Join(serveDir, r.URL.Path)
		info("Requested file: %s\n", filename)
		if strings.HasSuffix(filename, ".html") {
			filename = filename[0:len(filename)-5] + ".md"
		}
		if strings.HasSuffix(filename, "md") {
			if content, err := ioutil.ReadFile(filename); err == nil {
				if html, err := compile(content); err == nil {
					info("  Serve converted .md file.\n")
					w.Write(html)
					return
				}
			}
		}
		if r.URL.Path == "/favicon.ico" {
			info("  Serve internal favicon.ico.\n")
			w.Write(favIcon)
			return
		}
		info("  Serve raw file.\n")
		http.FileServer(http.Dir(serveDir)).ServeHTTP(w, r)
	})

	port := availablePort()
	info("start serving '%s' folder to localhost:%s.\n", serveDir, port)
	err := browser.Open("http://localhost:" + port + "/" + serveFile)
	try(err, "Can't open the web browser.")
	check(http.ListenAndServe("localhost:"+port, nil))
}

func main() {

	// error handling
	defer end()

	// The flags
	SetParameters()

	// serve or build ?
	if serve {
		serveFiles()
	} else {
		buildFiles()
	}
}
