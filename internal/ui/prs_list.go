package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

type pullRequestsListModel struct {
	prs []*gh.UserPullRequestsPullRequest

	list         list.Model
	delegateKeys pullRequestsListDelegateKeyMap

	selectedUser       string
	selectedOwner      string
	selectedRepository string
	width, height      int
}

type pullRequestsListDelegateKeyMap struct {
	open key.Binding
	back key.Binding
	quit key.Binding
}

func newPullRequestsListDelegateKeyMap() pullRequestsListDelegateKeyMap {
	return pullRequestsListDelegateKeyMap{
		open: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "open in browser"),
		),
		back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
		quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func newPullRequestsListModel() *pullRequestsListModel {
	delegateKeys := newPullRequestsListDelegateKeyMap()
	delegate := newPullRequestsListDelegate(delegateKeys)

	l := list.New(nil, delegate, 0, 0)
	l.KeyMap.Quit = delegateKeys.quit
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)

	return &pullRequestsListModel{
		list:         l,
		delegateKeys: delegateKeys,
	}
}

func (m *pullRequestsListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-2)
}

func (m *pullRequestsListModel) SetUser(id string) {
	m.selectedUser = id
}

func (m *pullRequestsListModel) setOwner(name string) {
	m.selectedOwner = name
}

func (m *pullRequestsListModel) setRepository(name string) {
	m.selectedRepository = name
}

func (m *pullRequestsListModel) updateList(prs []*gh.UserPullRequestsPullRequest) {
	m.prs = prs
	items := make([]list.Item, len(m.prs))
	for i, pr := range m.prs {
		created := formatDuration(pr.CretaedAt)
		closed := formatDuration(pr.ClosedAt)
		item := pullRequestsListItem{
			title:     pr.Title,
			status:    pr.State,
			number:    pr.Number,
			additions: pr.Additions,
			deletions: pr.Deletions,
			comments:  pr.Comments,
			created:   created,
			closed:    closed,
			url:       pr.Url,
		}
		items[i] = item
	}
	m.list.SetItems(items)
}

func (m pullRequestsListModel) Init() tea.Cmd {
	return nil
}

func (m pullRequestsListModel) openPullRequestPageInBrowser(item pullRequestsListItem) tea.Cmd {
	return func() tea.Msg {
		if err := openBrowser(item.url); err != nil {
			return profileErrorMsg{err, "failed to open browser"}
		}
		return nil
	}
}

func (m pullRequestsListModel) Update(msg tea.Msg) (pullRequestsListModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.delegateKeys.open):
			item := m.list.SelectedItem().(pullRequestsListItem)
			return m, m.openPullRequestPageInBrowser(item)
		case key.Matches(msg, m.delegateKeys.back):
			if m.list.FilterState() != list.Filtering {
				return m, goBackPullRequestsRepositoryPage
			}
		}
	case selectPullRequestsRepositoryMsg:
		m.list.ResetSelected()
		m.updateList(msg.repo.PullRequests)
		m.setRepository(msg.repo.Name)
		m.setOwner(msg.owner)
		return m, nil
	}
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pullRequestsListModel) View() string {
	return titleView(m.breadcrumb()) + listView(m.list)
}

func (m pullRequestsListModel) breadcrumb() []string {
	return []string{m.selectedUser, "PRs", m.selectedOwner, m.selectedRepository}
}
