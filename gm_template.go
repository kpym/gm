package main

import (
	_ "embed"
)

// defaultHTMLTemplate is the default value for `html` flag
var defaultHTMLTemplate string = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    {{- range .css }}
      {{- with .Url }}
        <link rel="stylesheet" type="text/css" href="{{.}}">
      {{- end }}
      {{- with .Code }}
        {{.}}
      {{- end }}
    {{- end }}
    {{- with .title }}
    <title>{{.}}</title>
    {{- end }}
  </head>
  <body>
    <article class="markdown-body">
{{.html}}
    </article>
    {{- if .liveupdate }}
    <script src="live.js#html,css"></script>
    {{- end }}
  </body>
</html>
`

// the favicon image for all served pages
//
//go:embed md.png
var favIcon []byte

// modified live.js script to serve locally
//
//go:embed live.js
var livejs []byte
