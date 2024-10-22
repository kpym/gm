# gm (a goldmark cli tool)

A cli tool converting Markdown to HTML.
This tool is a thin wrapper around the [github.com/yuin/goldmark](https://github.com/yuin/goldmark) library.


## Usage

### Single md to html

```
> gm file.md
```

### Open (and watch) md file in the browser

```
> gm --serve file.md
```


### The usage message

```
> gm -h
gm (version: 0.19.1): a goldmark cli tool which is a thin wrapper around github.com/yuin/goldmark (versio: v1.7.8).

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

  -s, --serve                    Start serving local .md file(s). No html is saved.
      --timeout int              Timeout in seconds for stop serving if no (non static) request. Default is 0 (no timeout).
  -c, --css stringArray          A css content or url or the theme name present in github.com/kpym/markdown-css. Multiple values are allowed. (default [github])
  -t, --title string             The page title. If empty, search for <h1> in the resulting html.
      --html string              The html template (file or string).
  -o, --out-dir string           The build output folder (created if not already existing, not used when serving).
      --readme-index             Compile README.md to index.html (not used when serving).
      --move-no-md               Move all non markdown non dot files to the output folder (not used when serving).
      --skip-dot                 Skip dot files (not used when serving).
      --pages                    Shortcut for --outdir='public' --readme-index --move-no-md --skip-dot (not used when serving).
      --links-md2html            Replace .md with .html in links to local files (not used when serving). (default true)
      --gm-attribute             goldmark option: allows to define attributes on some elements. (default true)
      --gm-auto-heading-id       goldmark option: enables auto heading ids. (default true)
      --gm-definition-list       goldmark option: enables definition lists. (default true)
      --gm-footnote              goldmark option: enables footnotes. (default true)
      --gm-linkify               goldmark option: activates auto links. (default true)
      --gm-strikethrough         goldmark option: enables strike through. (default true)
      --gm-table                 goldmark option: enables tables. (default true)
      --gm-task-list             goldmark option: enables task lists. (default true)
      --gm-typographer           goldmark option: activate punctuations substitution with typographic entities. (default true)
      --gm-emoji                 goldmark option: enables (github) emojis 💪. (default true)
      --gm-unsafe                goldmark option: enables raw html. (default true)
      --gm-hard-wraps            goldmark option: render newlines as <br>.
      --gm-xhtml                 goldmark option: render as XHTML.
      --gm-highlighting string   goldmark option: the code highlighting theme (empty string to disable).
                                 Check github.com/alecthomas/chroma for theme names. (default "github")
      --gm-line-numbers          goldmark option: enable line numering for code highlighting.
  -q, --quiet                    No errors and no info is printed. Return error code is still available.
  -h, --help                     Print this help message.
```

### How to

For more usage information check the [HOWTO](HOWTO.md) documentation.

## Installation

### Precompiled executables

You can download the executable for your platform from the [Releases](https://github.com/kpym/gm/releases).

### Compile it yourself

#### Using Go

```
$ go install github.com/kpym/gm@latest
```

#### Using goreleaser

After cloning this repo you can compile the sources with [goreleaser](https://github.com/goreleaser/goreleaser/) for all available platforms:

```
git clone https://github.com/kpym/gm.git .
goreleaser --snapshot --skip-publish --clean
```

You will find the resulting binaries in the `dist/` sub-folder.

## License

[MIT](LICENSE)
