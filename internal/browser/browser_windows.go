package browser

import (
	"os/exec"
)

// open the browser
func Open(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}
