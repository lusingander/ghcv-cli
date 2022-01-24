package ui

import (
	"errors"
	"os/exec"
	"runtime"
	"time"

	"github.com/ymotongpoo/datemaki"
)

func openBrowser(url string) error {
	if runtime.GOOS != "darwin" {
		return errors.New("unsupported os :(")
	}
	cmd := exec.Command("open", url)
	return cmd.Start()
}

func formatDuration(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	now := time.Now()
	return datemaki.FormatDurationFrom(now, t)
}
