package ui

import "github.com/charmbracelet/lipgloss"

const (
	appTitle = "GHCV"
)

var (
	titleBarStyle = lipgloss.NewStyle().
			Padding(0, 0, 1, 2)

	titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("97")).
			Foreground(lipgloss.Color("229")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Padding(1, 0, 0, 2)
)

var (
	selectedColor1 = lipgloss.Color("142")
	selectedColor2 = lipgloss.Color("143")
)

func titleView() string {
	// bubbles/list/styles.go
	title := titleStyle.Render(appTitle)
	return titleBarStyle.Render(title)
}
