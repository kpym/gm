package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kpym/gm/internal/browser"
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
	var lastMethodPath string
	var ticker *time.Ticker      // used to exit if no request is received for timeout seconds
	var exit atomic.Bool         // is set to true on every request and reset to false on every tick of the ticker
	var livejsactive atomic.Bool // is set to true if the last request was for non static content

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// say that we are alive
		exit.Store(false)
		// by default live.js is active (serving .md or related files)
		// it is reset to false if the request is for static content later
		livejsactive.Store(true)
		// how should I print the info?
		filename := filepath.Join(serveDir, r.URL.Path)
		newMethodPath := fmt.Sprintf("\n%s '%s':", r.Method, r.URL.Path)
		if newMethodPath != lastMethodPath {
			lastMethodPath = newMethodPath
			info(newMethodPath)
		}
		// serve the file
		if strings.HasSuffix(filename, ".html") {
			// try first to serve the corresponding .md file
			// if it is not present, serve the .html as static file
			filename = filename[0:len(filename)-5] + ".md"
		}
		if strings.HasSuffix(filename, "md") {
			if r.Method == "HEAD" {
				info(".")
				if fstat, err := os.Stat(filename); err == nil {
					w.Header().Set("Last-Modified", fstat.ModTime().UTC().Format(http.TimeFormat))
					w.Header().Set("Content-Type", "text/html")
					w.Write([]byte{})
				}
				return
			}
			if inFile, err := os.Open(filename); err == nil {
				defer inFile.Close()

				// convert the file to io.Reader
				var content io.Reader = inFile

				// Wrap the input with sedMdEngine if available
				if sedMdEngine != nil {
					content = sedMdEngine.Wrap(content)
				}

				markdown, err := io.ReadAll(content)
				if err != nil {
					check(err, "Problem reading the markdown.")
				}

				if html, err := compile(markdown); err == nil {
					// Wrap the HTML output with sedHtmlEngine if available
					if sedHtmlEngine != nil {
						htmlReader := sedHtmlEngine.Wrap(strings.NewReader(string(html)))
						html, err = io.ReadAll(htmlReader)
						if err != nil {
							check(err, "Problem applying sed-html commands.")
						}
					}

					info(" serve converted .md file.")
					w.Write(html)
					return
				}
			}
		}
		switch r.URL.Path {
		case "/favicon.ico":
			info(" serve internal png.")
			w.Header().Set("Cache-Control", "max-age=86400") // 86400 s = 1 day
			w.Header().Set("Expires", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))
			w.Write(favIcon)
			return
		case "/live.js":
			info(" serve live.js.")
			w.Header().Set("Cache-Control", "max-age=86400") // 86400 s = 1 day
			w.Header().Set("Expires", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))
			w.Write(livejs)
			return
		}
		info(" serve raw file.")
		livejsactive.Store(false) // is serving file without live.js
		w.Header().Set("Cache-Control", "no-store")
		http.FileServer(http.Dir(serveDir)).ServeHTTP(w, r)
	})

	// start the exit timer ?
	if liveupdate && timeout > 0 {
		// if no livejs request is received in timeout seconds, exit
		ticker = time.NewTicker(time.Duration(timeout) * time.Second)
		defer ticker.Stop()
		go func() {
			for {
				// wait for timeout seconds
				<-ticker.C
				// check if the last request was more than timeout seconds ago
				// and if it was for non static content, so that live.js should be active
				if exit.Swap(true) && livejsactive.Load() {
					info("\nNo request for %d seconds. Exit.\n\n", timeout)
					mainEnd()
				}
			}
		}()
	}

	port := availablePort()
	info("start serving '%s' folder to localhost:%s.\n", serveDir, port)
	url := "http://localhost:" + port + "/" + serveFile
	err := browser.Open(url)
	try(err, "Can't open the web browser, but you can visit now:", url)
	check(http.ListenAndServe("localhost:"+port, nil))
}
