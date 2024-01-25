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
- `{{.css}}` contains the css link obtained by the `--css` parameter;
- `{{.title}}` contains the first `h1` title, or the `--title` parameter if no `h1` title is present in the code.

```shell
> gm --html mymodel.html README.md
```

We can use a file or a string as `--html` parameter (run in bash here):

```shell
> echo "*test*" | gm -t "Test page" -c air --html $'title: {{.title}}\ncss: {{.css}}\nhtml: {{.html}}'
title: Test page
css: https://kpym.github.io/markdown-css/air.min.css
html: <p><em>test</em></p>
```

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
    - wget -c https://github.com/kpym/gm/releases/download/v0.16.0/gm_0.16.0_Linux_64bit.tar.gz -O - | tar -C /usr/local/bin -xz gm
    - gm --pages '**/*'
  artifacts:
    paths:
      - public
```
