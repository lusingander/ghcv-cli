package ui

import (
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

type pullRequestsListAllModel struct {
	prs *gh.UserPullRequests

	list         list.Model
	delegateKeys pullRequestsListAllDelegateKeyMap

	selectedUser  string
	width, height int
}

type pullRequestsListAllDelegateKeyMap struct {
	open key.Binding
	back key.Binding
	tog  key.Binding
	quit key.Binding
}

func newPullRequestsListAllDelegateKeyMap() pullRequestsListAllDelegateKeyMap {
	return pullRequestsListAllDelegateKeyMap{
		open: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "open in browser"),
		),
		back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
		tog: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle"),
		),
		quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func newPullRequestsListAllModel() *pullRequestsListAllModel {
	delegateKeys := newPullRequestsListAllDelegateKeyMap()
	delegate := newPullRequestsListAllDelegate(delegateKeys)

	l := list.New(nil, delegate, 0, 0)
	l.KeyMap.Quit = delegateKeys.quit
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)

	return &pullRequestsListAllModel{
		list:         l,
		delegateKeys: delegateKeys,
	}
}

func (m *pullRequestsListAllModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-2)
}

func (m *pullRequestsListAllModel) SetUser(id string) {
	m.selectedUser = id
}

func (m *pullRequestsListAllModel) updatePrs(prs *gh.UserPullRequests) {
	m.prs = prs
	items := make([]list.Item, 0)
	for _, owner := range m.prs.Owners {
		for _, repo := range owner.Repositories {
			for _, pr := range repo.PullRequests {
				created := formatDuration(pr.CretaedAt)
				closed := formatDuration(pr.ClosedAt)
				item := pullRequestsListAllItem{
					owner:      owner.Name,
					repository: repo.Name,
					createdAt:  pr.CretaedAt,
					closedAt:   pr.ClosedAt,
					pullRequestsListItem: pullRequestsListItem{
						title:     pr.Title,
						status:    pr.State,
						number:    pr.Number,
						additions: pr.Additions,
						deletions: pr.Deletions,
						comments:  pr.Comments,
						created:   created,
						closed:    closed,
						url:       pr.Url,
					},
				}
				items = append(items, item)
			}
		}
	}
	m.list.SetItems(items)
	m.sortItems()
}

func (m *pullRequestsListAllModel) sortItems() {
	items := m.list.Items()
	sort.Slice(items, func(i, j int) bool {
		return items[i].(pullRequestsListAllItem).createdAt.After(items[j].(pullRequestsListAllItem).createdAt)
	})
	m.list.SetItems(items)
}

func (m pullRequestsListAllModel) Init() tea.Cmd {
	return nil
}

func (m pullRequestsListAllModel) openPullRequestPageInBrowser(item pullRequestsListAllItem) tea.Cmd {
	return func() tea.Msg {
		if err := openBrowser(item.url); err != nil {
			return profileErrorMsg{err, "failed to open browser"}
		}
		return nil
	}
}

func (m pullRequestsListAllModel) Update(msg tea.Msg) (pullRequestsListAllModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.delegateKeys.open):
			item := m.list.SelectedItem().(pullRequestsListAllItem)
			return m, m.openPullRequestPageInBrowser(item)
		case key.Matches(msg, m.delegateKeys.back):
			if m.list.FilterState() != list.Filtering {
				return m, goBackMenuPage
			}
		case key.Matches(msg, m.delegateKeys.tog):
			return m, togglePullRequestsList
		}
	case togglePullRequestsListAllMsg:
		m.list.ResetSelected()
		m.updatePrs(msg.prs)
		return m, nil
	}
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pullRequestsListAllModel) View() string {
	return titleView(m.breadcrumb()) + listView(m.list)
}

func (m pullRequestsListAllModel) breadcrumb() []string {
	return []string{m.selectedUser, "PRs (ALL)"}
}
