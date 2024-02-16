package browser

import (
	"os/exec"
)

// open the browser
func Open(url string) (err error) {
	return exec.Command("xdg-open", url).Start()
}
