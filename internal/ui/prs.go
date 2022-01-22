package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

var (
	pullRequestsSpinnerStyle = lipgloss.NewStyle().
					Padding(2, 0, 0, 2)

	pullRequestsErrorStyle = lipgloss.NewStyle().
				Padding(2, 0, 0, 2).
				Foreground(lipgloss.Color("161"))
)

type pullRequestsInnerPage int

const (
	pullRequestsOwnerPage pullRequestsInnerPage = iota
	pullRequestsRepositoriesPage
	pullRequestsListPage
)

type pullRequestsModel struct {
	client      *gh.GitHubClient
	currentPage pullRequestsInnerPage

	owner   *pullRequestsOwnerModel
	spinner *spinner.Model

	errorMsg      *pullRequestsErrorMsg
	loading       bool
	width, height int
}

func newPullRequestsModel(client *gh.GitHubClient, s *spinner.Model) pullRequestsModel {
	return pullRequestsModel{
		client:  client,
		owner:   newPullRequestsOwnerModel(client),
		spinner: s,
	}
}

func (m *pullRequestsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.owner.SetSize(width, height)
}

func (m pullRequestsModel) Init() tea.Cmd {
	return nil
}

type pullRequestsSuccessMsg struct {
	prs *gh.UserPullRequests
}

var _ tea.Msg = (*pullRequestsSuccessMsg)(nil)

type pullRequestsErrorMsg struct {
	e       error
	summary string
}

var _ tea.Msg = (*pullRequestsErrorMsg)(nil)

type loadpullRequestsMsg struct{}

var _ tea.Msg = (*loadpullRequestsMsg)(nil)

func (m pullRequestsModel) loadPullRequests(id string) tea.Cmd {
	return func() tea.Msg {
		prs, err := m.client.QueryUserPullRequests(id)
		if err != nil {
			return pullRequestsErrorMsg{err, "failed to fetch pull requests"}
		}
		return pullRequestsSuccessMsg{prs}
	}
}

func (m *pullRequestsModel) updatePrs(prs *gh.UserPullRequests) {
	m.owner.updatePrs(prs)
}

func (m pullRequestsModel) Update(msg tea.Msg) (pullRequestsModel, tea.Cmd) {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case selectPullRequestsPageMsg:
		m.loading = true
		return m, m.loadPullRequests(msg.id)
	case pullRequestsSuccessMsg:
		m.errorMsg = nil
		m.loading = false
		m.currentPage = pullRequestsOwnerPage
		m.updatePrs(msg.prs)
		return m, nil
	case pullRequestsErrorMsg:
		m.errorMsg = &msg
		m.loading = false
		return m, nil
	}

	switch m.currentPage {
	case pullRequestsOwnerPage:
		*m.owner, cmd = m.owner.Update(msg)
		cmds = append(cmds, cmd)
	case pullRequestsRepositoriesPage:
		cmds = append(cmds, cmd)
	case pullRequestsListPage:
		cmds = append(cmds, cmd)
	default:
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m pullRequestsModel) View() string {
	if m.loading {
		return m.loadingView()
	}
	if m.errorMsg != nil {
		return m.errorView()
	}

	switch m.currentPage {
	case pullRequestsOwnerPage:
		return m.owner.View()
	case pullRequestsRepositoriesPage:
		return ""
	case pullRequestsListPage:
		return ""
	default:
		return ":("
	}
}

func (m pullRequestsModel) loadingView() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView()
	ret += title
	height -= cn(title)

	sp := pullRequestsSpinnerStyle.Render(m.spinner.View() + " Loading...")
	ret += sp
	height -= cn(sp)

	return ret
}

func (m pullRequestsModel) errorView() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView()
	ret += title
	height -= cn(title)

	errorText := pullRequestsErrorStyle.Render("ERROR: " + m.errorMsg.summary)
	ret += errorText
	height -= cn(errorText)

	return ret
}
