package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

var (
	pullRequestsErrorStyle = lipgloss.NewStyle().
		Padding(2, 0, 0, 2).
		Foreground(lipgloss.Color("161"))
)

type pullRequestsInnerPage int

const (
	pullRequestsOwnerPage pullRequestsInnerPage = iota
	pullRequestsRepositoryPage
	pullRequestsListPage
)

type pullRequestsModel struct {
	client      *gh.GitHubClient
	currentPage pullRequestsInnerPage

	owner   *pullRequestsOwnerModel
	repo    *pullRequestsRepositoryModel
	list    *pullRequestsListModel
	spinner *spinner.Model

	errorMsg      *pullRequestsErrorMsg
	loading       bool
	width, height int
}

func newPullRequestsModel(client *gh.GitHubClient, s *spinner.Model) pullRequestsModel {
	return pullRequestsModel{
		client:  client,
		owner:   newPullRequestsOwnerModel(client),
		repo:    newPullRequestsRepositoryModel(),
		list:    newPullRequestsListModel(),
		spinner: s,
	}
}

func (m *pullRequestsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.owner.SetSize(width, height)
	m.repo.SetSize(width, height)
	m.list.SetSize(width, height)
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

func (m pullRequestsModel) loadPullRequests(id string) tea.Cmd {
	return func() tea.Msg {
		prs, err := m.client.QueryUserPullRequests(id)
		if err != nil {
			return pullRequestsErrorMsg{err, "failed to fetch pull requests"}
		}
		return pullRequestsSuccessMsg{prs}
	}
}

type selectPullRequestsOwnerMsg struct {
	owner *gh.UserPullRequestsOwner
}

var _ tea.Msg = (*selectPullRequestsOwnerMsg)(nil)

type selectPullRequestsRepositoryMsg struct {
	repo *gh.UserPullRequestsRepository
}

var _ tea.Msg = (*selectPullRequestsRepositoryMsg)(nil)

type goBackPullRequestsOwnerPageMsg struct{}

var _ tea.Msg = (*goBackPullRequestsOwnerPageMsg)(nil)

func goBackPullRequestsOwnerPage() tea.Msg {
	return goBackPullRequestsOwnerPageMsg{}
}

type goBackPullRequestsRepositoryPageMsg struct{}

var _ tea.Msg = (*goBackPullRequestsRepositoryPageMsg)(nil)

func goBackPullRequestsRepositoryPage() tea.Msg {
	return goBackPullRequestsRepositoryPageMsg{}
}

func (m pullRequestsModel) Update(msg tea.Msg) (pullRequestsModel, tea.Cmd) {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case selectPullRequestsPageMsg:
		m.loading = true
		return m, m.loadPullRequests(msg.id)
	case selectPullRequestsOwnerMsg:
		m.currentPage = pullRequestsRepositoryPage
	case selectPullRequestsRepositoryMsg:
		m.currentPage = pullRequestsListPage
	case goBackPullRequestsOwnerPageMsg:
		m.currentPage = pullRequestsOwnerPage
	case goBackPullRequestsRepositoryPageMsg:
		m.currentPage = pullRequestsRepositoryPage
	case pullRequestsSuccessMsg:
		m.errorMsg = nil
		m.loading = false
		m.currentPage = pullRequestsOwnerPage
	case pullRequestsErrorMsg:
		m.errorMsg = &msg
		m.loading = false
		return m, nil
	}

	switch m.currentPage {
	case pullRequestsOwnerPage:
		*m.owner, cmd = m.owner.Update(msg)
		cmds = append(cmds, cmd)
	case pullRequestsRepositoryPage:
		*m.repo, cmd = m.repo.Update(msg)
		cmds = append(cmds, cmd)
	case pullRequestsListPage:
		*m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	default:
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m pullRequestsModel) View() string {
	if m.loading {
		return loadingView(m.height, m.spinner)
	}
	if m.errorMsg != nil {
		return m.errorView()
	}

	switch m.currentPage {
	case pullRequestsOwnerPage:
		return m.owner.View()
	case pullRequestsRepositoryPage:
		return m.repo.View()
	case pullRequestsListPage:
		return m.list.View()
	default:
		return baseStyle.Render("error... :(")
	}
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
