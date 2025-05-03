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
    - wget -c https://github.com/kpym/gm/releases/download/v0.23.0/gm_0.23.0_Linux_intel64.tar.gz -O - | tar -C /usr/local/bin -xz gm
    - gm --pages '**/*'
  artifacts:
    paths:
      - public
```

## Apply sed commands to markdown or HTML

The `--sed-md` and `--sed-html` flags allow you to apply `sed` commands to the markdown source or the resulting HTML output, respectively. These commands can be provided as inline strings or as files containing `sed` scripts. If the flags are used multiple times, the commands are combined into a single `sed` engine.

### Example: Modify markdown source

To replace all occurrences of `TODO` with `DONE` in the markdown source before conversion:

```shell
> gm --sed-md "s/TODO/DONE/g" file.md
```

### Example: Modify HTML output

To modify the resulting HTML code classes by replacing `language-` with `language `:

```shell
> gm --sed-html "s/class=\"language-/class=\"language /g" file.md
```

You can also use a file containing `sed` commands:

```shell
> gm --sed-html sed_commands.txt file.md
```

In this case, `sed_commands.txt` should contain the `sed` commands to be applied.
