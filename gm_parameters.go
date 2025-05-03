package main

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	chroma "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/spf13/pflag"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
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
	header := `gm (version: %s): a goldmark cli tool which is a thin wrapper around github.com/yuin/goldmark (version: %s).

  If not serving (no '--serve' or '-s' option is used):
  - the .md files are converted and saved as .html with the same base name;
  - if the corresponding .html file already exists, it is overwritten;
  - 'stdin' is converted to 'stdout';
  - when a pattern is used, only the matched .md files are considered.
  - the pattern can contain '*', '?', the '**' glob pattern, '[class]' and {alt1,...} alternatives.

  When serving (with '--serve' or '-s' option):
  - the .md files are converted and served as html with live.js (for live updates);
  - all other files are staticly served;
  - nothing is written on the disk.

`
	fmt.Fprintf(out, header, version, goldmarkVersion)
	pflag.PrintDefaults()
	fmt.Fprintf(out, "\n")
}

var (
	// serve flags
	serve     bool
	serveDir  string
	serveFile string
	timeout   int

	// build flags
	outdir     string
	inpatterns []string
	readme     bool
	move       bool
	skipdot    bool
	pages      bool

	// template flags
	css        []string
	title      string
	favicon    string
	htmlshell  string
	liveupdate bool

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
	emojis         bool
	unsafe         bool
	autoHeadingId  bool
	hardWraps      bool
	xhtml          bool
	chromatheme    string
	chromalines    bool
	// The following goldmark options are missing:
	// - Subscript / Superscript
	// - Ins
	// - Mark
	// - Inline footnote
	// - Compact style definition lists
	// - Abbreviations
	// - Math rendering

	mdParser goldmark.Markdown

	// post GoldMark flags
	localmdlinks bool

	// info flags
	quiet    bool
	showhelp bool
	info     func(string, ...interface{})

	// sed flags
	sedMd   []string
	sedHtml []string
)

// SetParameters configure the global variables from the command line flags.
func SetParameters() {
	pflag.BoolVarP(&serve, "serve", "s", false, "Start serving local .md file(s). No html is saved.")
	pflag.Lookup("serve").NoOptDefVal = "true"
	pflag.IntVar(&timeout, "timeout", 0, "Timeout in seconds for stop serving if no (non static) request. Default is 0 (no timeout).")

	pflag.StringArrayVarP(&css, "css", "c", []string{"github"}, "A css content or url or the theme name present in github.com/kpym/markdown-css. Multiple values are allowed.")
	pflag.StringVarP(&title, "title", "t", "", "The page title. If empty, search for <h1> in the resulting html.")
	pflag.StringVar(&favicon, "icon", "", "The favicon url.")
	pflag.StringVar(&htmlshell, "html", "", "The html template (file or string).")

	pflag.StringVarP(&outdir, "out-dir", "o", "", "The build output folder (created if not already existing, not used when serving).")
	pflag.BoolVar(&readme, "readme-index", false, "Compile README.md to index.html (not used when serving).")
	pflag.BoolVar(&move, "move-no-md", false, "Move all non markdown non dot files to the output folder (not used when serving).")
	pflag.BoolVar(&skipdot, "skip-dot", false, "Skip dot files (not used when serving).")
	pflag.BoolVar(&pages, "pages", false, "Shortcut for --outdir='public' --readme-index --move-no-md --skip-dot (not used when serving).")
	pflag.BoolVar(&localmdlinks, "links-md2html", true, "Replace .md with .html in links to local files (not used when serving).")

	pflag.BoolVar(&attribute, "gm-attribute", true, "goldmark option: allows to define attributes on some elements.")
	pflag.BoolVar(&autoHeadingId, "gm-auto-heading-id", true, "goldmark option: enables auto heading ids.")
	pflag.BoolVar(&definitionList, "gm-definition-list", true, "goldmark option: enables definition lists.")
	pflag.BoolVar(&footnote, "gm-footnote", true, "goldmark option: enables footnotes.")
	pflag.BoolVar(&linkify, "gm-linkify", true, "goldmark option: activates auto links.")
	pflag.BoolVar(&strikethrough, "gm-strikethrough", true, "goldmark option: enables strike through.")
	pflag.BoolVar(&table, "gm-table", true, "goldmark option: enables tables.")
	pflag.BoolVar(&taskList, "gm-task-list", true, "goldmark option: enables task lists.")
	pflag.BoolVar(&typographer, "gm-typographer", true, "goldmark option: activate punctuations substitution with typographic entities.")
	pflag.BoolVar(&emojis, "gm-emoji", true, "goldmark option: enables (github) emojis 💪.")
	pflag.BoolVar(&unsafe, "gm-unsafe", true, "goldmark option: enables raw html.")

	pflag.BoolVar(&hardWraps, "gm-hard-wraps", false, "goldmark option: render newlines as <br>.")
	pflag.BoolVar(&xhtml, "gm-xhtml", false, "goldmark option: render as XHTML.")

	pflag.StringVar(&chromatheme, "gm-highlighting", "github", "goldmark option: the code highlighting theme (empty string to disable).\nCheck github.com/alecthomas/chroma for theme names.")
	pflag.BoolVar(&chromalines, "gm-line-numbers", false, "goldmark option: enable line numering for code highlighting.")

	pflag.StringArrayVar(&sedMd, "sed-md", []string{}, "Apply sed file or command on the markdown source before conversion.")
	pflag.StringArrayVar(&sedHtml, "sed-html", []string{}, "Apply sed file or command on the HTML output after conversion.")

	pflag.BoolVarP(&quiet, "quiet", "q", false, "No errors and no info is printed. Return error code is still available.")
	pflag.BoolVarP(&showhelp, "help", "h", false, "Print this help message.")
	// keep the flags order
	pflag.CommandLine.SortFlags = false
	// in case of error do not display second time
	pflag.CommandLine.Init("gm", pflag.ContinueOnError)
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
	for i, c := range css {
		// if not empty and not containing '/' or '.' or '{' it should be a theme name
		if c != "" && !strings.ContainsAny(c, "/.{") {
			css[i] = "https://kpym.github.io/markdown-css/" + c + ".min.css"
		} else if strings.Contains(c, "{") {
			css[i] = "<style>" + c + "</style>"
		}
	}
	// set the template
	t, err := os.ReadFile(htmlshell)
	if err == nil {
		htmlshell = string(t)
	}
	if htmlshell == "" {
		htmlshell = defaultHTMLTemplate
	}

	//set flags from shortcuts
	if pages {
		if outdir == "" {
			outdir = "public"
		}
		readme = true
		move = true
		skipdot = true
	}

	// Initialize sed engines
	if err := addSedCommands(&sedMdEngine, sedMd); err != nil {
		check(err, "Failed to initialize sed-md engine.")
	}
	if err := addSedCommands(&sedHtmlEngine, sedHtml); err != nil {
		check(err, "Failed to initialize sed-html engine.")
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
		check(errors.New("only one file or folder can be specified for serving"))
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
		check(fmt.Errorf("the specified path '%s'is not a file or folder", filename))
	}

	// insert live.js in the template
	liveupdate = true
}

// setBuildParameters get all patterns and create (if necessary) the "out dir".
func setBuildParameters() {
	// get the positional parameters
	inpatterns = pflag.Args()
	// check for positional parameters
	if len(inpatterns) == 0 {
		// check if there is a pipeed input
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			inpatterns = append(inpatterns, "stdin")
		} else {
			pflag.Usage()
			check(errors.New("at least one input 'file.md', 'p*ttern' or 'stdin' should be provided"))
		}
	}

	// check the "out dir"
	if outdir != "" {
		outdir = filepath.Clean(outdir)
		if os.MkdirAll(outdir, os.ModePerm) != nil {
			check(fmt.Errorf("the specified output folder '%s' is not reachable", outdir))
		}
	}
}

// setGoldMark creates a new markdown parser with configuration based on the parameter flags.
// The code is borrowed from: https://github.com/gohugoio/hugo/blob/d90e37e0c6e812f9913bf256c9c81aa05b7a08aa/markup/goldmark/convert.go
func setGoldMark() {
	var (
		goldmarkOptions []goldmark.Option
		rendererOptions []renderer.Option
		extensions      []goldmark.Extender
		parserOptions   []parser.Option
	)

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
	if emojis {
		extensions = append(extensions, emoji.Emoji)
	}
	if attribute {
		parserOptions = append(parserOptions, parser.WithAttribute())
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

	if chromatheme != "" {
		var chromaOptions []chroma.Option
		chromaOptions = append(chromaOptions, chroma.WithLineNumbers(chromalines))

		extensions = append(extensions, highlighting.NewHighlighting(
			highlighting.WithStyle(chromatheme),
			highlighting.WithFormatOptions(chromaOptions...),
		))
	}

	goldmarkOptions = append(
		goldmarkOptions,
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(parserOptions...),
		goldmark.WithRendererOptions(rendererOptions...),
	)

	mdParser = goldmark.New(goldmarkOptions...)
}

// setTemplate parse the `html` flag to template.
func setTemplate() {
	var err error
	mdTemplate, err = template.New("md").Parse(htmlshell)
	check(err, "Problem parsing the HTML template.")
}
