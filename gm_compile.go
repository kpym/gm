package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/grokify/html-strip-tags-go"
)

// regexTitle is used to find the first h1 title (if any)
var regexTitle = regexp.MustCompile(`(?m)<h1[^>]*>(.*?)</h1>`)

// getTitle search for the first h1 title in the markdown
// if there is no one it returns the default title
func getTitle(htmlStr string) string {
	res := regexTitle.FindStringSubmatch(htmlStr)
	if len(res) > 1 {
		return strip.StripTags(string(res[1]))
	}

	return "GoldMark"
}

// compile convert markdown to full html
// by first applying markdown
// and then integrating the result in a html template
func compile(markdown []byte) (html []byte, err error) {
	// temporary buffer
	var htmlBuf bytes.Buffer

	// convert md to html code
	err = mdParser.Convert(markdown, &htmlBuf)
	if err != nil {
		return nil, fmt.Errorf("problem parsing markdown code to html with goldmark: %w", err)
	}
	htmlStr := htmlBuf.String()

	// combine the template and the resulting html
	var data = make(map[string]any)
	if title != "" {
		data["title"] = template.HTML(title)
	} else {
		data["title"] = template.HTML(getTitle(htmlStr))
	}
	// the css can be either an url or a code
	type cssType struct {
		Url  template.HTML
		Code template.HTML
	}
	cssall := make([]cssType, len(css))
	for i, c := range css {
		if strings.HasPrefix(c, "<style>") {
			cssall[i] = cssType{Code: template.HTML(c)}
		} else {
			cssall[i] = cssType{Url: template.HTML(c)}
		}
	}
	data["css"] = cssall
	data["html"] = template.HTML(htmlStr)
	if liveupdate {
		data["liveupdate"] = template.HTML("yes")
	}

	htmlBuf.Reset()
	err = mdTemplate.Execute(&htmlBuf, data)
	if err != nil {
		return nil, fmt.Errorf("problem building HTML from template: %w", err)
	}

	return htmlBuf.Bytes(), nil
}

// regexMdLink is used to identify .md links like href="xxxx.md"
var regexMdLink = regexp.MustCompile(`href\s*=\s*"([^"]+)\.md"`)

// replaceLinks replace all links like href="path/xxxx.md" to href="path/xxxx.html"
// if the file `path/xxxx.md` exists
func replaceLinks(html []byte, dir string) []byte {
	// replace .md links with .html for local files
	return regexMdLink.ReplaceAllFunc(html, func(s []byte) []byte {
		filename := strings.Split(string(s), `"`)[1]
		relname := filepath.Join(dir, filename)
		if _, err := os.Stat(relname); err != nil {
			return s
		}
		return []byte(fmt.Sprintf(`href="%s.html"`, filename[:len(filename)-3]))
	})
}
