package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
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

	userSelect   userSelectModel
	menu         menuModel
	repositories repositoriesModel

	spinner *spinner.Model
}

func newModel(client *gh.GitHubClient) model {
	s := spinner.New()
	s.Spinner = spinner.Moon
	return model{
		client:       client,
		currentPage:  userSelectPage,
		userSelect:   newUserSelectModel(client, &s),
		menu:         newMenuModel(),
		repositories: newRepositoriesModel(client, &s),
		spinner:      &s,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

type userSelectMsg struct {
	id string
}

var _ tea.Msg = (*userSelectMsg)(nil)

func userSelected(id string) tea.Cmd {
	return func() tea.Msg { return userSelectMsg{id} }
}

type selectRepositoriesPageMsg struct {
	id string
}

var _ tea.Msg = (*selectRepositoriesPageMsg)(nil)

func selectRepositoriesPage(id string) tea.Cmd {
	return func() tea.Msg { return selectRepositoriesPageMsg{id} }
}

type goBackUserSelectPageMsg struct{}

var _ tea.Msg = (*goBackUserSelectPageMsg)(nil)

func goBackUserSelectPage() tea.Msg {
	return goBackUserSelectPageMsg{}
}

type goBackMenuPageMsg struct{}

var _ tea.Msg = (*goBackMenuPageMsg)(nil)

func goBackMenuPage() tea.Msg {
	return goBackMenuPageMsg{}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit
		}
	case spinner.TickMsg:
		*m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		top, right, bottom, left := baseStyle.GetMargin()
		width := msg.Width - left - right
		height := msg.Height - top - bottom
		m.menu.SetSize(width, height)
		m.userSelect.SetSize(width, height)
		m.repositories.SetSize(width, height)
	case userSelectMsg:
		m.menu.selectedUser = msg.id
		m.currentPage = menuPage
		return m, nil
	case selectRepositoriesPageMsg:
		m.currentPage = repositoriesPage
	case goBackUserSelectPageMsg:
		m.userSelect.Reset()
		m.currentPage = userSelectPage
	case goBackMenuPageMsg:
		m.currentPage = menuPage
	}

	switch m.currentPage {
	case userSelectPage:
		m.userSelect, cmd = m.userSelect.Update(msg)
		cmds = append(cmds, cmd)
	case menuPage:
		m.menu, cmd = m.menu.Update(msg)
		cmds = append(cmds, cmd)
	case repositoriesPage:
		m.repositories, cmd = m.repositories.Update(msg)
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
	case repositoriesPage:
		return baseStyle.Render(m.repositories.View())
	}
	return baseStyle.Render("error... :(")
}

func Start(client *gh.GitHubClient) error {
	m := newModel(client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	return p.Start()
}
