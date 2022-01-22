package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

type pullRequestsOwnerModel struct {
	prs *gh.UserPullRequests

	list         list.Model
	delegateKeys pullRequestsOwnerDelegateKeyMap

	width, height int
}

type pullRequestsOwnerItem struct {
	name       string
	reposCount int
	prsCount   int
}

var _ list.DefaultItem = (*pullRequestsOwnerItem)(nil)

func (i pullRequestsOwnerItem) Title() string {
	return i.name
}

func (i pullRequestsOwnerItem) Description() string {
	var p, r string
	if i.prsCount > 1 {
		p = fmt.Sprintf("%d pull requests", i.prsCount)
	} else {
		p = "1 pull request"
	}
	if i.reposCount > 1 {
		r = fmt.Sprintf("%d repositories", i.reposCount)
	} else {
		r = "1 repository"
	}
	return fmt.Sprintf("Total %s in %s", p, r)
}

func (i pullRequestsOwnerItem) FilterValue() string {
	return i.name
}

type pullRequestsOwnerDelegateKeyMap struct {
	back key.Binding
}

func newPullRequestsOwnerDelegateKeyMap() pullRequestsOwnerDelegateKeyMap {
	return pullRequestsOwnerDelegateKeyMap{
		back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
	}
}

func newPullRequestsOwnerModel(client *gh.GitHubClient) *pullRequestsOwnerModel {
	var items []list.Item
	delegate := list.NewDefaultDelegate()

	delegateKeys := newPullRequestsOwnerDelegateKeyMap()
	delegate.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{delegateKeys.back}
	}
	delegate.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{{delegateKeys.back}}
	}

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().Foreground(selectedColor1).BorderForeground(selectedColor2)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().Foreground(selectedColor2).BorderForeground(selectedColor2)
	l := list.New(items, delegate, 0, 0)
	l.Title = appTitle
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)

	return &pullRequestsOwnerModel{
		list:         l,
		delegateKeys: delegateKeys,
	}
}

func (m *pullRequestsOwnerModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

func (m *pullRequestsOwnerModel) updatePrs(prs *gh.UserPullRequests) {
	m.prs = prs
	items := make([]list.Item, len(m.prs.Owners))
	for i, owner := range m.prs.Owners {
		repos := owner.Repositories
		prsCount := 0
		for _, repo := range repos {
			prsCount += len(repo.PullRequests)
		}
		item := pullRequestsOwnerItem{
			name:       owner.Name,
			reposCount: len(repos),
			prsCount:   prsCount,
		}
		items[i] = item
	}
	m.list.SetItems(items)
}

func (m pullRequestsOwnerModel) Init() tea.Cmd {
	return nil
}

func (m pullRequestsOwnerModel) Update(msg tea.Msg) (pullRequestsOwnerModel, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.delegateKeys.back):
			if m.list.FilterState() != list.Filtering {
				return m, goBackMenuPage
			}
		}
	}

	list, lCmd := m.list.Update(msg)
	m.list = list
	cmds = append(cmds, lCmd)

	return m, tea.Batch(cmds...)
}

func (m pullRequestsOwnerModel) View() string {
	return m.list.View()
}
