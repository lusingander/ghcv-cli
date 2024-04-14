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

	urlTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Underline(true)
)

var (
	selectedColor1 = lipgloss.Color("142")
	selectedColor2 = lipgloss.Color("143")

	listNormalTitleColorStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	listNormalItemStyle = lipgloss.NewStyle().
				Padding(0, 0, 0, 2)

	listNormalTitleStyle = listNormalTitleColorStyle.Copy().
				Padding(0, 0, 0, 2)

	listNormalDescColorStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})

	listNormalDescStyle = listNormalDescColorStyle.Copy().
				Padding(0, 0, 0, 2)

	listSelectedTitleColorStyle = lipgloss.NewStyle().
					Foreground(selectedColor1)

	listSelectedItemStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(selectedColor2).
				Padding(0, 0, 0, 1)

	listSelectedTitleStyle = listSelectedItemStyle.Copy().
				Foreground(selectedColor1)

	listSelectedDescColorStyle = listSelectedTitleColorStyle.Copy().
					Foreground(selectedColor2)

	listSelectedDescStyle = listSelectedItemStyle.Copy().
				Foreground(selectedColor2)
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

func loadingView(s *spinner.Model, bc []string) string {
	ret := ""

	title := titleView(bc)
	ret += title

	sp := spinnerStyle.Render(s.View() + " Loading...")
	ret += sp

	return ret
}

func cn(view string) int {
	return strings.Count(view, "\n")
}
