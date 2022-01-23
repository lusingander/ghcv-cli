package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

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

	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 0, 0, 2)

	listStyle = lipgloss.NewStyle().
			MarginTop(1)

	spinnerStyle = lipgloss.NewStyle().
			Padding(2, 0, 0, 2)

	helpStyle = lipgloss.NewStyle().
			Padding(1, 0, 0, 2)
)

var (
	selectedColor1 = lipgloss.Color("142")
	selectedColor2 = lipgloss.Color("143")
)

func titleView(bcs []string) string {
	// bubbles/list/styles.go
	title := titleStyle.Render(appTitle)
	if bcs != nil {
		title += breadcrumbStyle.Render(strings.Join(bcs, " > "))
	}
	return titleBarStyle.Render(title)
}

func listView(l list.Model) string {
	return listStyle.Render(l.View())
}

func loadingView(height int, s *spinner.Model, bc []string) string {
	if height <= 0 {
		return ""
	}

	ret := ""
	height = height - 1

	title := titleView(bc)
	ret += title
	height -= cn(title)

	sp := spinnerStyle.Render(s.View() + " Loading...")
	ret += sp
	height -= cn(sp)

	return ret
}

func cn(view string) int {
	return strings.Count(view, "\n")
}
