package ui

import (
	"errors"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/lusingander/ghcv-cli/internal/ghcv"
	"github.com/simonhege/timeago"
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
	return timeago.English.FormatRelativeDuration(now.Sub(t))
}

func isOrganizationLogin(s string) bool {
	return strings.HasPrefix(s, "@")
}

func organigzationUrlFrom(s string) string {
	login := strings.TrimSpace(strings.TrimLeft(s, "@"))
	return ghcv.GitHubBaseUrl + login
}

func isUrl(s string) bool {
	_, err := url.Parse(s)
	return err == nil
}
