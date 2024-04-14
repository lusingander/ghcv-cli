package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lusingander/ghcv-cli/internal/gh"
	"github.com/lusingander/kasane"
)

var (
	pullRequestsListAllDialogBodyStyle = lipgloss.NewStyle().
						Padding(0, 2)

	pullRequestsListAllDialogStyle = lipgloss.NewStyle().
					BorderStyle(lipgloss.RoundedBorder())

	pullRequestsListAllDialogSelectedStyle = lipgloss.NewStyle().
						Foreground(selectedColor1)

	pullRequestsListAllDialogNotSelectedStyle = lipgloss.NewStyle()
)

type pullRequestListAllSortType int

const (
	pullRequestListAllSortByCreatedAtDesc pullRequestListAllSortType = iota
	pullRequestListAllSortByCreatedAtAsc
)

type pullRequestStatus struct {
	name  string
	count int
}

type pullRequestsListAllModel struct {
	prs *gh.UserPullRequests

	list                           list.Model
	originalItems                  []list.Item
	delegateKeys                   pullRequestsListAllDelegateKeyMap
	filterStatusDialogDelegateKeys pullRequestsListAllFilterStatusDialogDelegateKeyMap

	selectedUser  string
	width, height int

	pullRequestListAllSortType

	statuses           []*pullRequestStatus
	statusIdx          int
	statusDialogOpened bool
}

type pullRequestsListAllDelegateKeyMap struct {
	stat key.Binding
	open key.Binding
	back key.Binding
	tog  key.Binding
	quit key.Binding
}

func newPullRequestsListAllDelegateKeyMap() pullRequestsListAllDelegateKeyMap {
	return pullRequestsListAllDelegateKeyMap{
		stat: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "filter by status"),
		),
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

type pullRequestsListAllFilterStatusDialogDelegateKeyMap struct {
	next  key.Binding
	prev  key.Binding
	close key.Binding
}

func newPullRequestsListAllFilterStatusDialogDelegateKeyMap() pullRequestsListAllFilterStatusDialogDelegateKeyMap {
	return pullRequestsListAllFilterStatusDialogDelegateKeyMap{
		next: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "select next"),
		),
		prev: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "select prev"),
		),
		close: key.NewBinding(
			key.WithKeys("T", "esc", "enter"),
			key.WithHelp("T", "close dialog"),
		),
	}
}

func newPullRequestsListAllModel() *pullRequestsListAllModel {
	delegateKeys := newPullRequestsListAllDelegateKeyMap()
	delegate := newPullRequestsListAllDelegate(delegateKeys)
	filterStatusDialogDelegateKeys := newPullRequestsListAllFilterStatusDialogDelegateKeyMap()

	l := list.New(nil, delegate, 0, 0)
	l.KeyMap.Quit = delegateKeys.quit
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)

	return &pullRequestsListAllModel{
		list:                           l,
		delegateKeys:                   delegateKeys,
		filterStatusDialogDelegateKeys: filterStatusDialogDelegateKeys,
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
	statusesMap := make(map[string]int)
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
				statusesMap[pr.State] += 1
			}
		}
	}
	m.list.SetItems(items)
	m.originalItems = items
	m.sortItems()

	m.statuses = []*pullRequestStatus{
		{name: "All", count: len(items)},
		{name: "OPEN", count: statusesMap["OPEN"]},
		{name: "MERGED", count: statusesMap["MERGED"]},
		{name: "CLOSED", count: statusesMap["CLOSED"]},
	}
	m.statusIdx = 0
}

func (m *pullRequestsListAllModel) sortItems() {
	items := m.list.Items()
	switch m.pullRequestListAllSortType {
	case pullRequestListAllSortByCreatedAtDesc:
		sort.Slice(items, func(i, j int) bool {
			return items[i].(pullRequestsListAllItem).createdAt.After(items[j].(pullRequestsListAllItem).createdAt)
		})
	case pullRequestListAllSortByCreatedAtAsc:
		sort.Slice(items, func(i, j int) bool {
			return items[i].(pullRequestsListAllItem).createdAt.Before(items[j].(pullRequestsListAllItem).createdAt)
		})
	}
	m.list.SetItems(items)
}

func (m *pullRequestsListAllModel) updateStatusIdx(reverse bool) {
	n := len(m.statuses)
	if reverse {
		m.statusIdx = ((m.statusIdx-1)%n + n) % n
	} else {
		m.statusIdx = (m.statusIdx + 1) % n
	}
}

func (m *pullRequestsListAllModel) filterItems() {
	if m.statuses[m.statusIdx].name == "All" {
		m.list.SetItems(m.originalItems)
		m.sortItems()
		return
	}
	items := make([]list.Item, 0)
	for _, i := range m.originalItems {
		if i.(pullRequestsListAllItem).status == m.statuses[m.statusIdx].name {
			items = append(items, i)
		}
	}
	m.list.SetItems(items)
	m.sortItems()
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
		if m.statusDialogOpened {
			switch {
			case key.Matches(msg, m.filterStatusDialogDelegateKeys.close):
				m.statusDialogOpened = false
			case key.Matches(msg, m.filterStatusDialogDelegateKeys.next):
				m.list.ResetSelected()
				m.updateStatusIdx(false)
				m.filterItems()
			case key.Matches(msg, m.filterStatusDialogDelegateKeys.prev):
				m.list.ResetSelected()
				m.updateStatusIdx(true)
				m.filterItems()
			}
			return m, nil
		}
		switch {
		case key.Matches(msg, m.delegateKeys.stat):
			m.statusDialogOpened = true
			return m, nil
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
	ret := titleView(m.breadcrumb()) + listView(m.list)
	if m.statusDialogOpened {
		return m.withStatusDialogView(ret)
	}
	return ret
}

func (m pullRequestsListAllModel) withStatusDialogView(base string) string {
	title := repositoriesDialogTitleStyle.Render("Status")

	ivs := make([]string, len(m.statuses))
	for i, s := range m.statuses {
		ivs[i] = m.statusKeySelectItemView(s)
	}
	body := strings.Join(ivs, "\n")
	body = pullRequestsListAllDialogBodyStyle.Render(body)

	dialog := pullRequestsListAllDialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, body))

	dw, dh := lipgloss.Size(dialog)
	top := (m.height / 2) - (dh / 2)
	left := (m.width / 2) - (dw / 2)
	return kasane.OverlayString(base, dialog, top, left, kasane.WithPadding(m.width))
}

func (m pullRequestsListAllModel) statusKeySelectItemView(status *pullRequestStatus) string {
	if m.statuses[m.statusIdx].name == status.name {
		return pullRequestsListAllDialogSelectedStyle.Render(fmt.Sprintf("> %s (%d)", status.name, status.count))
	} else {
		return pullRequestsListAllDialogNotSelectedStyle.Render(fmt.Sprintf("  %s (%d)", status.name, status.count))
	}
}

func (m pullRequestsListAllModel) breadcrumb() []string {
	return []string{m.selectedUser, "PRs (ALL)"}
}
