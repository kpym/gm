package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kpym/goldmark-cli/internal/browser"
)

// availablePort provides the first available port after 8080
// or 8180 if no available ports are present.
func availablePort() (port string) {
	for i := 8080; i < 8181; i++ {
		port = strconv.Itoa(i)
		if ln, err := net.Listen("tcp", "localhost:"+port); err == nil {
			ln.Close()
			break
		}
	}
	return port
}

// serveFiles serve the local folder `serveDir`.
// If an .md (or corresponding .html) file is requested it is compiled and send as html.
func serveFiles() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filename := filepath.Join(serveDir, r.URL.Path)
		info("Requested file: %s\n", filename)
		if strings.HasSuffix(filename, ".html") {
			filename = filename[0:len(filename)-5] + ".md"
		}
		if strings.HasSuffix(filename, "md") {
			if content, err := ioutil.ReadFile(filename); err == nil {
				if html, err := compile(content); err == nil {
					info("  Serve converted .md file.\n")
					w.Write(html)
					return
				}
			}
		}
		if r.URL.Path == "/favicon.ico" {
			info("  Serve internal favicon.ico.\n")
			w.Write(favIcon)
			return
		}
		info("  Serve raw file.\n")
		http.FileServer(http.Dir(serveDir)).ServeHTTP(w, r)
	})

	port := availablePort()
	info("start serving '%s' folder to localhost:%s.\n", serveDir, port)
	url := "http://localhost:" + port + "/" + serveFile
	err := browser.Open(url)
	try(err, "Can't open the web browser, but you can visit now:", url)
	check(http.ListenAndServe("localhost:"+port, nil))
}
