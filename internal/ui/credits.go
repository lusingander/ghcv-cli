package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	creditsViewportStyle = lipgloss.NewStyle().
				Padding(1, 0, 1, 2)

	creditsRepositoryNameStyle = lipgloss.NewStyle().
					Bold(true)

	creditsUrlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Underline(true)

	creditsSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	creditsSeparator = "----------------------------------------"
)

type creditsKeyMap struct {
	Back key.Binding
	Quit key.Binding
}

func (k creditsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Back,
		k.Quit,
	}
}

func (k creditsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Back,
		},
		{
			k.Quit,
		},
	}
}

type creditsModel struct {
	viewport viewport.Model

	keys creditsKeyMap
	help help.Model

	width, height int
}

func newCreditsModel() creditsModel {
	keys := creditsKeyMap{
		Back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}

	contents := ""
	for _, c := range credits {
		contents += creditsRepositoryNameStyle.Render(c.name) + "\n\n"
		contents += creditsUrlStyle.Render(c.url) + "\n\n"
		contents += c.text + "\n"
		contents += creditsSeparatorStyle.Render(creditsSeparator) + "\n\n"
	}

	v := viewport.New(0, 0)
	v.SetContent(contents)

	return creditsModel{
		keys:     keys,
		viewport: v,
		help:     help.New(),
	}
}

func (m *creditsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.help.Width = width
	t, r, b, l := creditsViewportStyle.GetPadding()
	m.viewport.Width = width - r - l
	m.viewport.Height = height - 2 - t - b
}

func (m creditsModel) Init() tea.Cmd {
	return nil
}

func (m creditsModel) Update(msg tea.Msg) (creditsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, goBackHelpPage
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m creditsModel) View() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView(m.breadcrumb())
	ret += title
	height -= cn(title)

	credits := creditsViewportStyle.Render(m.viewport.View())
	ret += credits
	height -= cn(credits)

	help := helpStyle.Render(m.help.View(m.keys))
	height -= cn(help)

	ret += strings.Repeat("\n", height)
	ret += help

	return ret
}

func (m creditsModel) breadcrumb() []string {
	return []string{"Help", "Credits"}
}
