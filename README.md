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

Usage: gm [options] [file.md].
  If the markdown file is missing the standard input is used in place.
  The available options are:

  -s, --css string     The css file or the theme name present in github.com/kpym/markdown-css (default "github")
  -t, --title string   The page title.
  -h, --help           Print this help message.
```

### Piped input with parameters
```shell
> cat file.md | gm -t "Test page" -s jasonm23-markdown
```

Here `jasonm23-markdown` is converted to `https://kpym.github.io/markdown-css/jasonm23-markdown.min.css`.

## Installation

### Precompiled executables

You can download the executable for your platform from the [Realases](https://github.com/kpym/goldmark-cli/releases).

### Compile it yourself

#### Using Go

This method will comile to executable named `goldmark-cli` and not `gm`.

```shell
$ go get github.com/kpym/goldmark-cli
```

#### Using goreleaser

After cloning this repo you can comile the sources with [goreleaser](https://github.com/goreleaser/goreleaser/) for all available platforms:

```shell
git clone https://github.com/kpym/goldmark-cli.git .
goreleaser --snapshot --skip-publish --rm-dist
```

You will find the resulting binaries in the `dist/` subfolder.

## License

MIT
