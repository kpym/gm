# gm

Cli tool converting Markdown to HTML.
This tool is a thin wrapper around the [github.com/yuin/goldmark](https://github.com/yuin/goldmark) library.


## Usage

### Simple md to html
```shell
> gm file.md > file.html
```

### The usage message
```shell
> ./gm -h
gm (version: --): a goldmark cli tool which is a thin wrapper around github.com/yuin/goldmark.

Usage: gm [options] <file.md|file pattern|stdin>.

  If a file pattern is used, only the mached .md files are used.
  The .md files are converted to .html with the same name.
  If the .html file exists it is overwritten.
  The available options are:

  -s, --css string        The css file or the theme name present in github.com/kpym/markdown-css (default "github")
  -t, --title string      The page title.
      --html string       The html shell (file or string).
      --attribute         Allows to define attributes on some elements. (default true)
      --auto-heading-id   Enables auto heading ids. (default true)
      --definition-list   Enables definition lists. (default true)
      --footnote          Enables footnotes. (default true)
      --linkify           Activates auto links. (default true)
      --strikethrough     Enables strike through. (default true)
      --table             Enables tables. (default true)
      --task-list         Enables task lists. (default true)
      --typographer       Activate punctuations substitution with typographic entities. (default true)
      --unsafe            Enables raw html. (default true)
      --hard-wraps        Render newlines as <br>.
      --xhtml             Render as XHTML.
      --links-md2html     Convert links to local .md files to corresponding .html. (default true)
  -h, --help              Print this help message.

```

### Piped input with parameters
```shell
> cat file.md | gm -t "Test page" -s jasonm23-markdown
```

Here `jasonm23-markdown` is converted to `https://kpym.github.io/markdown-css/jasonm23-markdown.min.css`.

### List of the available themes

The list off all available css themre is: `air`, `github`, `jasonm23-dark`, `jasonm23-foghorn`, `jasonm23-markdown`, `jasonm23-swiss`, `markedapp-byword`, `mixu-page`, `mixu-radar`, `modest`, `retro`, `roryg-ghostwriter`, `splendor`, `thomasf-solarizedcssdark`, `thomasf-solarizedcsslight`, `witex`.

All this theme are hosted on GitHub pages of the [markdown-css](https://github.com/kpym/markdown-css) project.

### Custom HTML template

The custom HTML template can contain the following variables:

- `{{.html}}` contains the parsed html code from the markdown
- `{{.css}}` contains the css link obtained by the `--css` parameter
- `{{.title}}` contains title string obtained by the `--title` parameter

```shell
> gm --html mymodel.html README.md
```

We can use a file or a string as `--html` parameter (run in bash here):

```shell
> echo "*test*" | gm -t "Test page" -s air --html $'title: {{.title}}\ncss: {{.css}}\nhtml: {{.html}}'
title: Test page
css: https://kpym.github.io/markdown-css/air.min.css
html: <p><em>test</em></p>
```

## Installation

### Precompiled executables

You can download the executable for your platform from the [Releases](https://github.com/kpym/goldmark-cli/releases).

### Compile it yourself

#### Using Go

This method will compile to executable named `goldmark-cli` and not `gm`.

```shell
$ go get github.com/kpym/goldmark-cli
```

#### Using goreleaser

After cloning this repo you can compile the sources with [goreleaser](https://github.com/goreleaser/goreleaser/) for all available platforms:

```shell
git clone https://github.com/kpym/goldmark-cli.git .
goreleaser --snapshot --skip-publish --rm-dist
```

You will find the resulting binaries in the `dist/` subfolder.

## License

MIT
