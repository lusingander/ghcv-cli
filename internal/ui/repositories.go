package ui

import (
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lusingander/ghcv-cli/internal/gh"
	"github.com/lusingander/kasane"
)

var (
	repositoriesErrorStyle = lipgloss.NewStyle().
				Padding(2, 0, 0, 2).
				Foreground(lipgloss.Color("161"))

	dialogTitleStyle = lipgloss.NewStyle().
				Align(lipgloss.Center).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color("240"))

	repositoriesSortDialogTitleStyle = dialogTitleStyle.Copy().
						Width(30)

	repositoriesSortDialogBodyStyle = lipgloss.NewStyle().
					Padding(0, 2)

	reposirotiesSortDialogStyle = lipgloss.NewStyle().
					BorderStyle(lipgloss.RoundedBorder())

	repositoriesSortDialogSelectedStyle = lipgloss.NewStyle().
						Foreground(selectedColor1)

	repositoriesSortDialogNotSelectedStyle = lipgloss.NewStyle()
)

type sortType int

const (
	sortByStarDesc sortType = iota
	sortByStarAsc
	sortByUpdatedDesc
	sortByUpdatedAsc
)

type repositoriesModel struct {
	client *gh.GitHubClient

	list    list.Model
	spinner *spinner.Model

	delegateKeys           repositoriesDelegateKeyMap
	sortDialogDelegateKeys repositoriesSortDialogDelegateKeyMap

	errorMsg      *repositoriesErrorMsg
	loading       bool
	selectedUser  string
	width, height int

	sortType
	sortDialogOpened bool
}

type repositoriesDelegateKeyMap struct {
	sort key.Binding
	open key.Binding
	back key.Binding
	quit key.Binding
}

func newRepositoriesDelegateKeyMap() repositoriesDelegateKeyMap {
	return repositoriesDelegateKeyMap{
		sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort"),
		),
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

type repositoriesSortDialogDelegateKeyMap struct {
	next  key.Binding
	prev  key.Binding
	close key.Binding
}

func newRepositoriesSortDialogDelegateKeyMap() repositoriesSortDialogDelegateKeyMap {
	return repositoriesSortDialogDelegateKeyMap{
		next: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "select next"),
		),
		prev: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "select prev"),
		),
		close: key.NewBinding(
			key.WithKeys("s", "esc", "enter"),
			key.WithHelp("s", "close dialog"),
		),
	}
}

func newRepositoriesModel(client *gh.GitHubClient, s *spinner.Model) repositoriesModel {
	delegateKeys := newRepositoriesDelegateKeyMap()
	delegate := NewRepositoryDelegate(delegateKeys)
	sortDialogDelegateKeys := newRepositoriesSortDialogDelegateKeyMap()

	l := list.New(nil, delegate, 0, 0)
	l.KeyMap.Quit = delegateKeys.quit
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)

	return repositoriesModel{
		client:                 client,
		list:                   l,
		spinner:                s,
		delegateKeys:           delegateKeys,
		sortDialogDelegateKeys: sortDialogDelegateKeys,
	}
}

func (m *repositoriesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-2)
}

func (m *repositoriesModel) SetUser(id string) {
	m.selectedUser = id
}

func (m *repositoriesModel) updateItems(repos *gh.UserRepositories) {
	items := make([]list.Item, len(repos.Repositories))
	for i, repo := range repos.Repositories {
		updated := formatDuration(repo.PushedAt)
		item := &repositoryItem{
			title:       repo.Name,
			description: repo.Description,
			langName:    repo.LangName,
			langColor:   repo.LangColor,
			license:     repo.License,
			updated:     updated,
			stars:       repo.Stars,
			forks:       repo.Forks,
			watchers:    repo.Watchers,
			url:         repo.Url,
			pushedAt:    repo.PushedAt,
		}
		items[i] = item
	}
	m.list.SetItems(items)
	m.sortType = sortByStarDesc
}

func (m *repositoriesModel) updateSortType(reverse bool) {
	if reverse {
		switch m.sortType {
		case sortByStarDesc:
			m.sortType = sortByUpdatedAsc
		case sortByStarAsc:
			m.sortType = sortByStarDesc
		case sortByUpdatedDesc:
			m.sortType = sortByStarAsc
		case sortByUpdatedAsc:
			m.sortType = sortByUpdatedDesc
		}
	} else {
		switch m.sortType {
		case sortByStarDesc:
			m.sortType = sortByStarAsc
		case sortByStarAsc:
			m.sortType = sortByUpdatedDesc
		case sortByUpdatedDesc:
			m.sortType = sortByUpdatedAsc
		case sortByUpdatedAsc:
			m.sortType = sortByStarDesc
		}
	}
}

func (m *repositoriesModel) sortItems() {
	items := m.list.Items()
	switch m.sortType {
	case sortByStarDesc:
		sort.Slice(items, func(i, j int) bool {
			return items[i].(*repositoryItem).stars > items[j].(*repositoryItem).stars
		})
	case sortByStarAsc:
		sort.Slice(items, func(i, j int) bool {
			return items[i].(*repositoryItem).stars < items[j].(*repositoryItem).stars
		})
	case sortByUpdatedDesc:
		sort.Slice(items, func(i, j int) bool {
			return items[i].(*repositoryItem).pushedAt.After(items[j].(*repositoryItem).pushedAt)
		})
	case sortByUpdatedAsc:
		sort.Slice(items, func(i, j int) bool {
			return items[i].(*repositoryItem).pushedAt.Before(items[j].(*repositoryItem).pushedAt)
		})
	}
	m.list.SetItems(items)
}

func (m repositoriesModel) Init() tea.Cmd {
	return nil
}

type repositoriesSuccessMsg struct {
	repos *gh.UserRepositories
}

var _ tea.Msg = (*repositoriesSuccessMsg)(nil)

type repositoriesErrorMsg struct {
	e       error
	summary string
}

var _ tea.Msg = (*repositoriesErrorMsg)(nil)

type loadRepositoriesMsg struct{}

var _ tea.Msg = (*loadRepositoriesMsg)(nil)

func (m repositoriesModel) loadRepositores(id string) tea.Cmd {
	return func() tea.Msg {
		repos, err := m.client.QueryUserRepositories(id)
		if err != nil {
			return repositoriesErrorMsg{err, "failed to fetch repositories"}
		}
		return repositoriesSuccessMsg{repos}
	}
}

func (m repositoriesModel) openRepositoryPageInBrowser(item *repositoryItem) tea.Cmd {
	return func() tea.Msg {
		if err := openBrowser(item.url); err != nil {
			return profileErrorMsg{err, "failed to open browser"}
		}
		return nil
	}
}

func (m repositoriesModel) Update(msg tea.Msg) (repositoriesModel, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		if m.sortDialogOpened {
			switch {
			case key.Matches(msg, m.sortDialogDelegateKeys.close):
				m.sortDialogOpened = false
			case key.Matches(msg, m.sortDialogDelegateKeys.next):
				m.list.ResetSelected()
				m.updateSortType(false)
				m.sortItems()
			case key.Matches(msg, m.sortDialogDelegateKeys.prev):
				m.list.ResetSelected()
				m.updateSortType(true)
				m.sortItems()
			}
			return m, nil
		}
		switch {
		case key.Matches(msg, m.delegateKeys.sort):
			m.sortDialogOpened = true
			return m, nil
		case key.Matches(msg, m.delegateKeys.open):
			item := m.list.SelectedItem().(*repositoryItem)
			return m, m.openRepositoryPageInBrowser(item)
		case key.Matches(msg, m.delegateKeys.back):
			if m.list.FilterState() != list.Filtering {
				return m, goBackMenuPage
			}
		}
	case selectRepositoriesPageMsg:
		m.loading = true
		return m, m.loadRepositores(msg.id)
	case repositoriesSuccessMsg:
		m.errorMsg = nil
		m.loading = false
		m.list.ResetSelected()
		m.updateItems(msg.repos)
		return m, nil
	case repositoriesErrorMsg:
		m.errorMsg = &msg
		m.loading = false
		return m, nil
	}

	list, lCmd := m.list.Update(msg)
	m.list = list
	cmds = append(cmds, lCmd)

	return m, tea.Batch(cmds...)
}

func (m repositoriesModel) View() string {
	if m.loading {
		return loadingView(m.height, m.spinner, m.breadcrumb())
	}
	if m.errorMsg != nil {
		return m.errorView()
	}
	ret := titleView(m.breadcrumb()) + listView(m.list)
	if m.sortDialogOpened {
		return m.withSortDialogView(ret)
	}
	return ret
}

func (m repositoriesModel) withSortDialogView(base string) string {
	title := repositoriesSortDialogTitleStyle.Render("Sort")

	body := ""
	body += m.sortKeySelectItemView("Stars (Desc)", sortByStarDesc)
	body += "\n"
	body += m.sortKeySelectItemView("Stars (Asc)", sortByStarAsc)
	body += "\n"
	body += m.sortKeySelectItemView("Last Updated (Desc)", sortByUpdatedDesc)
	body += "\n"
	body += m.sortKeySelectItemView("Last Updated (Asc)", sortByUpdatedAsc)
	body = repositoriesSortDialogBodyStyle.Render(body)

	dialog := reposirotiesSortDialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, body))

	dw, dh := lipgloss.Size(dialog)
	top := (m.height / 2) - (dh / 2)
	left := (m.width / 2) - (dw / 2)
	return kasane.OverlayString(base, dialog, top, left, kasane.WithPadding(m.width))
}

func (m repositoriesModel) sortKeySelectItemView(s string, st sortType) string {
	if m.sortType == st {
		return repositoriesSortDialogSelectedStyle.Render("> " + s)
	} else {
		return repositoriesSortDialogNotSelectedStyle.Render("  " + s)
	}
}

func (m repositoriesModel) errorView() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView(m.breadcrumb())
	ret += title
	height -= cn(title)

	errorText := repositoriesErrorStyle.Render("ERROR: " + m.errorMsg.summary)
	ret += errorText
	height -= cn(errorText)

	return ret
}

func (m repositoriesModel) breadcrumb() []string {
	return []string{m.selectedUser, "Repositories"}
}
