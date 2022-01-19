package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

var (
	baseStyle = lipgloss.NewStyle().Margin(1, 2)
)

type page int

const (
	authPage page = iota
	userSelectPage
	menuPage
	profilePage
	pullRequrstsPage
	repositoriesPage
)

type model struct {
	client      *gh.GitHubClient
	currentPage page

	selectedUser string

	userSelect userSelectModel
	menu       menuModel
}

func newModel(client *gh.GitHubClient) model {
	return model{
		client:      client,
		currentPage: userSelectPage,
		userSelect:  newUserSelectModel(client),
		menu:        newMenuModel(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.userSelect.spinner.Tick,
	)
}

type userSelectMsg struct {
	id string
}

var _ tea.Msg = (*userSelectMsg)(nil)

func userSelected(id string) tea.Cmd {
	return func() tea.Msg { return userSelectMsg{id} }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := baseStyle.GetMargin()
		width := msg.Width - left - right
		height := msg.Height - top - bottom
		m.menu.SetSize(width, height)
		m.userSelect.SetSize(width, height)
	case userSelectMsg:
		m.selectedUser = msg.id
		m.currentPage = menuPage
		return m, nil
	}

	switch m.currentPage {
	case userSelectPage:
		userSelect, cmd := m.userSelect.Update(msg)
		m.userSelect = userSelect
		cmds = append(cmds, cmd)
	case menuPage:
		menu, cmd := m.menu.Update(msg)
		m.menu = menu
		cmds = append(cmds, cmd)
	default:
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.currentPage {
	case userSelectPage:
		return baseStyle.Render(m.userSelect.View())
	case menuPage:
		return baseStyle.Render(m.menu.View())
	}
	return baseStyle.Render("error... :(")
}

func Start(client *gh.GitHubClient) error {
	m := newModel(client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	return p.Start()
}
