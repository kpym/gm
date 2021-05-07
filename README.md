# gm (a goldmark cli tool)

A cli tool converting Markdown to HTML.
This tool is a thin wrapper around the [github.com/yuin/goldmark](https://github.com/yuin/goldmark) library.


## Usage

### Single md to html

```shell
> gm file.md
```

### Open (and watch) md file in the browser

```shell
> gm --serve file.md
```


### The usage message

```shell
> gm -h
gm (version: --): a goldmark cli tool which is a thin wrapper around github.com/yuin/goldmark.

Usage: gm [options] (file.md|file pattern|stdin)+.

  If not serving (no `--serve` or `-s` option is used):
  - if  file pattern is used, only the matched .md files are used;
  - the .md files are converted to .html with the same name;
  - if the .html file exists it is overwritten.

  The available options are:

  -s, --serve                   Start serving local .md file(s). No html is saved.
  -c, --css string              A css url or the theme name present in github.com/kpym/markdown-css. (default "github")
  -t, --title string            The default page title. Used if no h1 is found in the .md file.
      --html string             The html template (file or string).
  -o, --out-dir string          The build output folder (created if not already existing, not used if --serve).
      --gm-attribute            goldmark option: allows to define attributes on some elements. (default true)
      --gm-auto-heading-id      goldmark option: enables auto heading ids. (default true)
      --gm-definition-list      goldmark option: enables definition lists. (default true)
      --gm-footnote             goldmark option: enables footnotes. (default true)
      --gm-linkify              goldmark option: activates auto links. (default true)
      --gm-strikethrough        goldmark option: enables strike through. (default true)
      --gm-table                goldmark option: enables tables. (default true)
      --gm-task-list            goldmark option: enables task lists. (default true)
      --gm-typographer          goldmark option: activate punctuations substitution with typographic entities. (default true)
      --gm-unsafe               goldmark option: enables raw html. (default true)
      --gm-hard-wraps           goldmark option: render newlines as <br>.
      --gm-xhtml                goldmark option: render as XHTML.
      --links-md2html --serve   Replace .md with .html in links to local files (not used if --serve). (default true)
  -q, --quiet                   No errors and no info is printed. Return error code is still available.
  -h, --help                    Print this help message.

```

### How to

For more usage information check the [HOWTO](HOWTO.md) documentation.

## Installation

### Precompiled executables

You can download the executable for your platform from the [Releases](https://github.com/kpym/gm/releases).

### Compile it yourself

#### Using Go

```shell
$ go get github.com/kpym/gm
```

#### Using goreleaser

After cloning this repo you can compile the sources with [goreleaser](https://github.com/goreleaser/goreleaser/) for all available platforms:

```shell
git clone https://github.com/kpym/gm.git .
goreleaser --snapshot --skip-publish --rm-dist
```

You will find the resulting binaries in the `dist/` sub-folder.

## License

[MIT](LICENSE)
