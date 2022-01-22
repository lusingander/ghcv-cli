package ui

import (
	"errors"
	"os/exec"
	"runtime"
)

func openBrowser(url string) error {
	if runtime.GOOS != "darwin" {
		return errors.New("unsupported os :(")
	}
	cmd := exec.Command("open", url)
	return cmd.Start()
}
