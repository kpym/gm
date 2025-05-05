# How to use gm (a goldmark cli tool).

## Convert multiple files at once

```shell
> gm '*.md' 'subfolder/*.md' -o outfolder
```

## Convert piped input with specified theme

```shell
> cat file.md | gm -c jasonm23-markdown stdin > output.html
```
Here `jasonm23-markdown` is converted to `https://kpym.github.io/markdown-css/jasonm23-markdown.min.css`.

### Available themes

The list off all available css themre is: `air`, `github`, `jasonm23-dark`, `jasonm23-foghorn`, `jasonm23-markdown`, `jasonm23-swiss`, `markedapp-byword`, `mixu-page`, `mixu-radar`, `modest`, `retro`, `roryg-ghostwriter`, `splendor`, `thomasf-solarizedcssdark`, `thomasf-solarizedcsslight`, `witex`.

All this theme are hosted on GitHub pages of the [markdown-css](https://github.com/kpym/markdown-css) project.

## Custom HTML template

The custom HTML template can contain the following variables:

- `{{.html}}` contains the parsed html code from the markdown;
- `{{.css}}` contains a list of css links or codes obtained from the `--css` parameter;
- `{{.title}}` contains the first `h1` title, or the `--title` parameter if no `h1` title is present in the code.

```shell
> gm --html mymodel.html README.md
```

We can use a file or a string as `--html` parameter (run in bash here):

```shell
> echo "*test*" | gm -q -t "Test page" -c air --html $'title: {{.title}}\ncss: {{.css}}\nhtml: {{.html}}'
title: Test page
css: [{https://kpym.github.io/markdown-css/air.min.css }]
html: <p><em>test</em></p>
```

The default template is [gm_template.html](gm_template.html).

## Serve at localhost

When used with `--serve`/`-s` flag `gm` start serving all files from the specified folder. The `.md` files are converted and served as `html` but all other files are staticly served. To serve the current folder you can simply run:

```shell
> gm -s
```

When some `.md` file is requested it is converted and wraped in a full `html` with `live.js` inside. In this way every second a `HEAD` request is made and if the file changes the reulting `html` is updated.

When we specify a file, like in
```shell
> gm -s some/folder/file.md
```
the served folder is `some/folder/` and the requested url is `localhost:8080/file.md` _(if `8080` is available)_.


## Use gm to produce a GitLab pages website

Here is an example of possible `.gitlab-ci.yml`:

```yaml
pages:
  image: alpine
  script:
    - wget -c https://github.com/kpym/gm/releases/download/v0.24.0/gm_0.24.0_Linux_intel64.tar.gz -O - | tar -C /usr/local/bin -xz gm
    - gm --pages '**/*'
  artifacts:
    paths:
      - public
```

## Apply regex substitutions to markdown or HTML

The `--re-md` and `--re-html` flags allow you to apply regex substitutions to the markdown source or the resulting HTML output, respectively. These substitutions can be provided as inline strings or as files containing regex rules (one rule per line). These flags can be used multiple times to apply multiple substitutions.

The replace rules are specified in the format 
```
  <delimiter><pattern><delimiter><replacement>[<delimiter>[<comment>]]
```
The delimiters can be any character, but `/` is commonly used. The pattern is a regular expression, and the replacement is the string to replace the matched pattern with. The optional comment can be used to document the rule. For example, the rule `/foo/bar/` replaces all occurrences of `foo` with `bar`. 

### Modify markdown source

To replace all occurrences of `TODO` with `DONE` in the markdown source before conversion:

```shell
> gm --re-md "/TODO/DONE/" file.md
```

### Modify HTML output

To modify the resulting HTML code classes by removing `language-` prefix from the class tag:

```shell
> gm --re-html "/class=\"language-/class=\"/" file.md
```

### Combining regex rules

You can specify multiple rules by using the flag multiple times:

```shell
> gm --re-html "/class=\"language-/class=\"/" --re-html "/foo/bar/" file.md
```

You can also use a file containing lines of regex rules:

```shell
> gm --re-html replace_rules.txt file.md
```

Where `replace_rules.txt` contains for example:
```
/class="language-/class="/
/foo/bar/
```

And you can combine both inline and file rules:

```shell
> gm --re-md "|TODO|DONE|" --re-html replace_rules.txt --re-html ";bad;good;" file.md
```