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
	helpPage
	aboutPage
	creditsPage
)

type model struct {
	client      *gh.GitHubClient
	currentPage page

	userSelect   userSelectModel
	menu         menuModel
	profile      profileModel
	pullRequests pullRequestsModel
	repositories repositoriesModel
	help         helpModel
	about        aboutModel
	credits      creditsModel

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
		profile:      newProfileModel(client, &s),
		pullRequests: newPullRequestsModel(client, &s),
		repositories: newRepositoriesModel(client, &s),
		help:         newHelpModel(),
		about:        newAboutModel(),
		credits:      newCreditsModel(),
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

type selectProfilePageMsg struct {
	id string
}

var _ tea.Msg = (*selectProfilePageMsg)(nil)

func selectProfilePage(id string) tea.Cmd {
	return func() tea.Msg { return selectProfilePageMsg{id} }
}

type goBackUserSelectPageMsg struct{}

var _ tea.Msg = (*goBackUserSelectPageMsg)(nil)

func goBackUserSelectPage() tea.Msg {
	return goBackUserSelectPageMsg{}
}

type selectPullRequestsPageMsg struct {
	id string
}

var _ tea.Msg = (*selectPullRequestsPageMsg)(nil)

func selectPullRequestsPage(id string) tea.Cmd {
	return func() tea.Msg { return selectPullRequestsPageMsg{id} }
}

type selectHelpPageMsg struct{}

var _ tea.Msg = (*selectHelpPageMsg)(nil)

func selectHelpPage() tea.Msg {
	return selectHelpPageMsg{}
}

type selectAboutPageMsg struct{}

var _ tea.Msg = (*selectAboutPageMsg)(nil)

func selectAboutPage() tea.Msg {
	return selectAboutPageMsg{}
}

type selectCreditsPageMsg struct{}

var _ tea.Msg = (*selectCreditsPageMsg)(nil)

func selectCreditsPage() tea.Msg {
	return selectCreditsPageMsg{}
}

type goBackMenuPageMsg struct{}

var _ tea.Msg = (*goBackMenuPageMsg)(nil)

func goBackMenuPage() tea.Msg {
	return goBackMenuPageMsg{}
}

type goBackHelpPageMsg struct{}

var _ tea.Msg = (*goBackHelpPageMsg)(nil)

func goBackHelpPage() tea.Msg {
	return goBackHelpPageMsg{}
}

func (m *model) SetSize(width, height int) {
	m.userSelect.SetSize(width, height)
	m.menu.SetSize(width, height)
	m.profile.SetSize(width, height)
	m.pullRequests.SetSize(width, height)
	m.repositories.SetSize(width, height)
	m.help.SetSize(width, height)
	m.about.SetSize(width, height)
	m.credits.SetSize(width, height)
}

func (m *model) SetUser(id string) {
	m.menu.SetUser(id)
	m.profile.SetUser(id)
	m.repositories.SetUser(id)
	m.pullRequests.SetUser(id)
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
		m.SetSize(msg.Width-left-right, msg.Height-top-bottom)
	case userSelectMsg:
		m.SetUser(msg.id)
		m.currentPage = menuPage
	case selectProfilePageMsg:
		m.currentPage = profilePage
	case selectPullRequestsPageMsg:
		m.currentPage = pullRequrstsPage
	case selectRepositoriesPageMsg:
		m.currentPage = repositoriesPage
	case selectHelpPageMsg:
		m.currentPage = helpPage
	case selectAboutPageMsg:
		m.currentPage = aboutPage
	case selectCreditsPageMsg:
		m.currentPage = creditsPage
	case goBackUserSelectPageMsg:
		m.currentPage = userSelectPage
	case goBackMenuPageMsg:
		m.currentPage = menuPage
	case goBackHelpPageMsg:
		m.currentPage = helpPage
	}

	switch m.currentPage {
	case userSelectPage:
		m.userSelect, cmd = m.userSelect.Update(msg)
		cmds = append(cmds, cmd)
	case menuPage:
		m.menu, cmd = m.menu.Update(msg)
		cmds = append(cmds, cmd)
	case profilePage:
		m.profile, cmd = m.profile.Update(msg)
		cmds = append(cmds, cmd)
	case pullRequrstsPage:
		m.pullRequests, cmd = m.pullRequests.Update(msg)
		cmds = append(cmds, cmd)
	case repositoriesPage:
		m.repositories, cmd = m.repositories.Update(msg)
		cmds = append(cmds, cmd)
	case helpPage:
		m.help, cmd = m.help.Update(msg)
		cmds = append(cmds, cmd)
	case aboutPage:
		m.about, cmd = m.about.Update(msg)
		cmds = append(cmds, cmd)
	case creditsPage:
		m.credits, cmd = m.credits.Update(msg)
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
	case profilePage:
		return baseStyle.Render(m.profile.View())
	case pullRequrstsPage:
		return baseStyle.Render(m.pullRequests.View())
	case repositoriesPage:
		return baseStyle.Render(m.repositories.View())
	case helpPage:
		return baseStyle.Render(m.help.View())
	case aboutPage:
		return baseStyle.Render(m.about.View())
	case creditsPage:
		return baseStyle.Render(m.credits.View())
	}
	return baseStyle.Render("error... :(")
}

func Start(client *gh.GitHubClient) error {
	m := newModel(client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
