package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

// exit is a shared variable used to stop the server.
type exit struct {
	// Locker is used to prevent concurrent access to the exit variable.
	Locker sync.RWMutex
	do     bool
}

func (e *exit) no() {
	e.Locker.Lock()
	e.do = false
	e.Locker.Unlock()
}

func (e *exit) yes() {
	e.Locker.Lock()
	e.do = true
	e.Locker.Unlock()
}

func (e *exit) isYes() bool {
	e.Locker.RLock()
	defer e.Locker.RUnlock()
	return e.do
}

// serveFiles serve the local folder `serveDir`.
// If an .md (or corresponding .html) file is requested it is compiled and send as html.
func serveFiles() {
	var lastMethodPath string
	// variables for exit on idle (for 2 seconds)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	exit := exit{} // the initial value is false

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// say thet we are alive
		exit.no()
		// how should I print the info?
		filename := filepath.Join(serveDir, r.URL.Path)
		newMethodPath := fmt.Sprintf("\n%s '%s':", r.Method, r.URL.Path)
		if newMethodPath != lastMethodPath {
			lastMethodPath = newMethodPath
			info(newMethodPath)
		}
		// serve the file
		if strings.HasSuffix(filename, ".html") {
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
			if content, err := os.ReadFile(filename); err == nil {
				if html, err := compile(content); err == nil {
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
		case "live.js":
			info(" serve live.js.")
			w.Header().Set("Cache-Control", "max-age=86400") // 86400 s = 1 day
			w.Header().Set("Expires", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))
			w.Write(livejs)
			return
		}
		info(" serve raw file.")
		w.Header().Set("Cache-Control", "no-store")
		http.FileServer(http.Dir(serveDir)).ServeHTTP(w, r)
	})

	if liveupdate {
		// start the exit timer
		// if no request is received in 2 seconds, exit
		go func() {
			for {
				// wait for 2 seconds
				<-ticker.C
				if exit.isYes() {
					info("\nNo request for 2 seconds. Exit.\n\n")
					mainEnd()
				}
				exit.yes() // should be rest to no by the next request in less than 2 seconds
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
