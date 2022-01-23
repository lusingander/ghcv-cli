package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

type pullRequestsRepositoryModel struct {
	repos []*gh.UserPullRequestsRepository

	list         list.Model
	delegateKeys pullRequestsRepositoryDelegateKeyMap

	selectedUser  string
	selectedOwner string
	width, height int
}

type pullRequestsRepositoryDelegateKeyMap struct {
	open key.Binding
	sel  key.Binding
	back key.Binding
	quit key.Binding
}

func newPullRequestsRepositoryDelegateKeyMap() pullRequestsRepositoryDelegateKeyMap {
	return pullRequestsRepositoryDelegateKeyMap{
		open: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "open in browser"),
		),
		sel: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
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

func newPullRequestsRepositoryModel() *pullRequestsRepositoryModel {
	delegateKeys := newPullRequestsRepositoryDelegateKeyMap()
	delegate := newPullRequestsRepositoryDelegate(delegateKeys)

	l := list.New(nil, delegate, 0, 0)
	l.KeyMap.Quit = delegateKeys.quit
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)

	return &pullRequestsRepositoryModel{
		list:         l,
		delegateKeys: delegateKeys,
	}
}

func (m *pullRequestsRepositoryModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-2)
}

func (m *pullRequestsRepositoryModel) SetUser(id string) {
	m.selectedUser = id
}

func (m *pullRequestsRepositoryModel) setOwner(name string) {
	m.selectedOwner = name
}

func (m *pullRequestsRepositoryModel) updateRepos(repos []*gh.UserPullRequestsRepository) {
	m.repos = repos
	items := make([]list.Item, len(m.repos))
	for i, repo := range m.repos {
		item := &pullRequestsRepositoryItem{
			name:        repo.Name,
			description: repo.Description,
			langName:    repo.LangName,
			langColor:   repo.LangColor,
			prsCount:    len(repo.PullRequests),
			url:         repo.Url,
		}
		items[i] = item
	}
	m.list.SetItems(items)
}

func (m pullRequestsRepositoryModel) Init() tea.Cmd {
	return nil
}

func (m pullRequestsRepositoryModel) selectPullRequestsRepository(name string) tea.Cmd {
	return func() tea.Msg {
		for _, repo := range m.repos {
			if repo.Name == name {
				return selectPullRequestsRepositoryMsg{repo, m.selectedOwner}
			}
		}
		return pullRequestsErrorMsg{nil, "failed to get repository"}
	}
}

func (m pullRequestsRepositoryModel) openRepositoryPageInBrowser(item *pullRequestsRepositoryItem) tea.Cmd {
	return func() tea.Msg {
		if err := openBrowser(item.url); err != nil {
			return profileErrorMsg{err, "failed to open browser"}
		}
		return nil
	}
}

func (m pullRequestsRepositoryModel) Update(msg tea.Msg) (pullRequestsRepositoryModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.delegateKeys.open):
			item := m.list.SelectedItem().(*pullRequestsRepositoryItem)
			return m, m.openRepositoryPageInBrowser(item)
		case key.Matches(msg, m.delegateKeys.sel):
			item := m.list.SelectedItem().(*pullRequestsRepositoryItem)
			return m, m.selectPullRequestsRepository(item.name)
		case key.Matches(msg, m.delegateKeys.back):
			if m.list.FilterState() != list.Filtering {
				return m, goBackPullRequestsOwnerPage
			}
		}
	case selectPullRequestsOwnerMsg:
		m.updateRepos(msg.owner.Repositories)
		m.setOwner(msg.owner.Name)
		return m, nil
	}
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pullRequestsRepositoryModel) View() string {
	return titleView(m.breadcrumb()) + listView(m.list)
}

func (m pullRequestsRepositoryModel) breadcrumb() []string {
	return []string{m.selectedUser, "PRs", m.selectedOwner}
}
