package main

import (
	_ "embed"
)

// defaultHTMLTemplate is the default value for `html` flag
//
//go:embed gm_template.html
var defaultHTMLTemplate string

// the favicon image for all served pages
//
//go:embed md.png
var favIcon []byte

// modified live.js script to serve locally
//
//go:embed live.js
var livejs []byte
