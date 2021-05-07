package browser

import (
	"os/exec"
)

// open the browser
func Open(url string) (err error) {
	if err = exec.Command("xdg-open", url).Start(); err == nil {
		return nil
	}
	// are we in WSL ?
	if exec.Command("cmd.exe", "/C", "start", url).Start() == nil {
		return nil
	}
	// return the xdg-open err
	return err
}
