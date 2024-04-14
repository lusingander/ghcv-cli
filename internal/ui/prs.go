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
	pullRequestsListAllPage
)

type pullRequestsModel struct {
	client      *gh.GitHubClient
	currentPage pullRequestsInnerPage

	prs *gh.UserPullRequests

	owner   *pullRequestsOwnerModel
	repo    *pullRequestsRepositoryModel
	list    *pullRequestsListModel
	listAll *pullRequestsListAllModel
	spinner *spinner.Model

	errorMsg      *pullRequestsErrorMsg
	loading       bool
	selectedUser  string
	width, height int
}

func newPullRequestsModel(client *gh.GitHubClient, s *spinner.Model) pullRequestsModel {
	return pullRequestsModel{
		client:  client,
		owner:   newPullRequestsOwnerModel(),
		repo:    newPullRequestsRepositoryModel(),
		list:    newPullRequestsListModel(),
		listAll: newPullRequestsListAllModel(),
		spinner: s,
	}
}

func (m *pullRequestsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.owner.SetSize(width, height)
	m.repo.SetSize(width, height)
	m.list.SetSize(width, height)
	m.listAll.SetSize(width, height)
}

func (m *pullRequestsModel) SetUser(id string) {
	m.selectedUser = id
	m.owner.SetUser(id)
	m.repo.SetUser(id)
	m.list.SetUser(id)
	m.listAll.SetUser(id)
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
	repo  *gh.UserPullRequestsRepository
	owner string
}

var _ tea.Msg = (*selectPullRequestsRepositoryMsg)(nil)

type togglePullRequestsListMsg struct{}

var _ tea.Msg = (*togglePullRequestsListMsg)(nil)

func togglePullRequestsList() tea.Msg {
	return togglePullRequestsListMsg{}
}

type togglePullRequestsListAllMsg struct {
	prs *gh.UserPullRequests
}

var _ tea.Msg = (*togglePullRequestsListAllMsg)(nil)

func togglePullRequestsListAll(prs *gh.UserPullRequests) tea.Cmd {
	return func() tea.Msg {
		return togglePullRequestsListAllMsg{prs}
	}
}

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
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
	case selectPullRequestsPageMsg:
		m.loading = true
		return m, m.loadPullRequests(msg.id)
	case selectPullRequestsOwnerMsg:
		m.currentPage = pullRequestsRepositoryPage
	case selectPullRequestsRepositoryMsg:
		m.currentPage = pullRequestsListPage
	case togglePullRequestsListMsg:
		m.currentPage = pullRequestsOwnerPage
	case togglePullRequestsListAllMsg:
		m.currentPage = pullRequestsListAllPage
	case goBackPullRequestsOwnerPageMsg:
		m.currentPage = pullRequestsOwnerPage
	case goBackPullRequestsRepositoryPageMsg:
		m.currentPage = pullRequestsRepositoryPage
	case pullRequestsSuccessMsg:
		m.errorMsg = nil
		m.loading = false
		m.prs = msg.prs
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
	case pullRequestsListAllPage:
		*m.listAll, cmd = m.listAll.Update(msg)
		cmds = append(cmds, cmd)
	default:
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m pullRequestsModel) View() string {
	if m.loading {
		return loadingView(m.spinner, m.breadcrumb())
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
	case pullRequestsListAllPage:
		return m.listAll.View()
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

	title := titleView(m.breadcrumb())
	ret += title
	height -= cn(title)

	errorText := pullRequestsErrorStyle.Render("ERROR: " + m.errorMsg.summary)
	ret += errorText
	height -= cn(errorText)

	return ret
}

func (m pullRequestsModel) breadcrumb() []string {
	return []string{m.selectedUser, "PRs"}
}
