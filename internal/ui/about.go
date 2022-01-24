package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lusingander/ghcv-cli/internal/ghcv"
)

var (
	aboutItemStyle = lipgloss.NewStyle().
			Padding(1, 0, 1, 2)

	aboutItemAppNameStyle = profileItemStyle.Copy().
				Bold(true)

	aboutItemUrlStyle = profileItemStyle.Copy().
				Foreground(lipgloss.Color("33")).
				Underline(true)
)

type aboutKeyMap struct {
	Open key.Binding
	Back key.Binding
	Quit key.Binding
}

func (k aboutKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Open,
		k.Back,
		k.Quit,
	}
}

func (k aboutKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Open,
		},
		{
			k.Back,
		},
		{
			k.Quit,
		},
	}
}

type aboutModel struct {
	keys aboutKeyMap
	help help.Model

	width, height int
}

func newAboutModel() aboutModel {
	keys := aboutKeyMap{
		Open: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "open in browser"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
	return aboutModel{
		keys: keys,
		help: help.New(),
	}
}

func (m *aboutModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.help.Width = width
}

func (m aboutModel) Init() tea.Cmd {
	return nil
}

func (m aboutModel) openThisRepositoryPageInBrowser() tea.Cmd {
	return func() tea.Msg {
		if err := openBrowser(ghcv.AppUrl); err != nil {
			return profileErrorMsg{err, "failed to open browser"}
		}
		return nil
	}
}

func (m aboutModel) Update(msg tea.Msg) (aboutModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Open):
			return m, m.openThisRepositoryPageInBrowser()
		case key.Matches(msg, m.keys.Back):
			return m, goBackHelpPage
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m aboutModel) View() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView(m.breadcrumb())
	ret += title
	height -= cn(title)

	appName := aboutItemAppNameStyle.Render(ghcv.AppName)
	ret += appName
	height -= cn(appName)

	ver := aboutItemStyle.Render("Version " + ghcv.Version)
	ret += ver
	height -= cn(ver)

	appUrl := aboutItemUrlStyle.Render(ghcv.AppUrl)
	ret += appUrl
	height -= cn(appUrl)

	help := helpStyle.Render(m.help.View(m.keys))
	height -= cn(help)

	ret += strings.Repeat("\n", height)
	ret += help

	return ret
}

func (m aboutModel) breadcrumb() []string {
	return []string{"Help", "About"}
}
