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

// the favicon image for all served pages
//go:embed md.png
var favIcon []byte
