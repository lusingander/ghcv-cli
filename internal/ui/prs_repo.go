package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

type pullRequestsRepositoryModel struct {
	repos []*gh.UserPullRequestsRepository

	list         list.Model
	delegateKeys pullRequestsRepositoryDelegateKeyMap

	width, height int
}

type pullRequestsRepositoryItem struct {
	name     string
	prsCount int
}

var _ list.DefaultItem = (*pullRequestsRepositoryItem)(nil)

func (i pullRequestsRepositoryItem) Title() string {
	return i.name
}

func (i pullRequestsRepositoryItem) Description() string {
	var p string
	if i.prsCount > 1 {
		p = fmt.Sprintf("%d pull requests", i.prsCount)
	} else {
		p = "1 pull request"
	}
	return fmt.Sprintf("Total %s", p)
}

func (i pullRequestsRepositoryItem) FilterValue() string {
	return i.name
}

type pullRequestsRepositoryDelegateKeyMap struct {
	open key.Binding
	back key.Binding
}

func newPullRequestsRepositoryDelegateKeyMap() pullRequestsRepositoryDelegateKeyMap {
	return pullRequestsRepositoryDelegateKeyMap{
		open: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
	}
}

func newPullRequestsRepositoryModel() *pullRequestsRepositoryModel {
	var items []list.Item
	delegate := list.NewDefaultDelegate()

	delegateKeys := newPullRequestsRepositoryDelegateKeyMap()
	delegate.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{delegateKeys.open, delegateKeys.back}
	}
	delegate.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{{delegateKeys.open, delegateKeys.back}}
	}

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().Foreground(selectedColor1).BorderForeground(selectedColor2)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().Foreground(selectedColor2).BorderForeground(selectedColor2)
	l := list.New(items, delegate, 0, 0)
	l.Title = appTitle
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)

	return &pullRequestsRepositoryModel{
		list:         l,
		delegateKeys: delegateKeys,
	}
}

func (m *pullRequestsRepositoryModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

func (m *pullRequestsRepositoryModel) updateRepos(repos []*gh.UserPullRequestsRepository) {
	m.repos = repos
	items := make([]list.Item, len(m.repos))
	for i, repo := range m.repos {
		item := pullRequestsRepositoryItem{
			name:     repo.Name,
			prsCount: len(repo.PullRequests),
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
				return selectPullRequestsRepositoryMsg{repo}
			}
		}
		return pullRequestsErrorMsg{nil, "failed to get repository"}
	}
}

func (m pullRequestsRepositoryModel) Update(msg tea.Msg) (pullRequestsRepositoryModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.delegateKeys.open):
			item := m.list.SelectedItem().(pullRequestsRepositoryItem)
			return m, m.selectPullRequestsRepository(item.name)
		case key.Matches(msg, m.delegateKeys.back):
			if m.list.FilterState() != list.Filtering {
				return m, goBackPullRequestsOwnerPage
			}
		}
	case selectPullRequestsOwnerMsg:
		m.updateRepos(msg.owner.Repositories)
		return m, nil
	}
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pullRequestsRepositoryModel) View() string {
	return m.list.View()
}
