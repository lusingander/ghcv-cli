package ui

import (
	"fmt"
	"sort"
	"strings"

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

	repositoriesDialogTitleStyle = dialogTitleStyle.Copy().
					Width(30)

	repositoriesDialogBodyStyle = lipgloss.NewStyle().
					Padding(0, 2)

	reposirotiesDialogStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder())

	repositoriesDialogSelectedStyle = lipgloss.NewStyle().
					Foreground(selectedColor1)

	repositoriesDialogNotSelectedStyle = lipgloss.NewStyle()
)

type sortType int

const (
	sortByStarDesc sortType = iota
	sortByStarAsc
	sortByUpdatedDesc
	sortByUpdatedAsc
)

type repositoeisLang struct {
	name  string
	count int
}

type repositoriesModel struct {
	client *gh.GitHubClient

	list          list.Model
	originalItems []list.Item
	spinner       *spinner.Model

	delegateKeys           repositoriesDelegateKeyMap
	sortDialogDelegateKeys repositoriesSortDialogDelegateKeyMap
	langDialogDelegateKeys repositoriesLangDialogDelegateKeyMap

	errorMsg      *repositoriesErrorMsg
	loading       bool
	selectedUser  string
	width, height int

	sortType
	sortDialogOpened bool

	langs            []*repositoeisLang
	langIdx          int
	langDialogOpened bool
}

type repositoriesDelegateKeyMap struct {
	sort key.Binding
	lang key.Binding
	open key.Binding
	back key.Binding
	quit key.Binding
}

func newRepositoriesDelegateKeyMap() repositoriesDelegateKeyMap {
	return repositoriesDelegateKeyMap{
		sort: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "sort"),
		),
		lang: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "filter by language"),
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
			key.WithKeys("S", "esc", "enter"),
			key.WithHelp("S", "close dialog"),
		),
	}
}

type repositoriesLangDialogDelegateKeyMap struct {
	next  key.Binding
	prev  key.Binding
	close key.Binding
}

func newRepositoriesLangDialogDelegateKeyMap() repositoriesLangDialogDelegateKeyMap {
	return repositoriesLangDialogDelegateKeyMap{
		next: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "select next"),
		),
		prev: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "select prev"),
		),
		close: key.NewBinding(
			key.WithKeys("L", "esc", "enter"),
			key.WithHelp("L", "close dialog"),
		),
	}
}

func newRepositoriesModel(client *gh.GitHubClient, s *spinner.Model) repositoriesModel {
	delegateKeys := newRepositoriesDelegateKeyMap()
	delegate := NewRepositoryDelegate(delegateKeys)
	sortDialogDelegateKeys := newRepositoriesSortDialogDelegateKeyMap()
	langDialogDelegateKeys := newRepositoriesLangDialogDelegateKeyMap()

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
		langDialogDelegateKeys: langDialogDelegateKeys,
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
	langMap := make(map[string]int)
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
		langMap[repo.LangName] += 1
	}

	m.list.SetItems(items)
	m.originalItems = items

	m.sortType = sortByStarDesc

	langs := make([]*repositoeisLang, 0, len(langMap)+1)
	langs = append(langs, &repositoeisLang{name: "All", count: len(items)})
	for k, v := range langMap {
		langs = append(langs, &repositoeisLang{name: k, count: v})
	}
	sort.Slice(langs, func(i, j int) bool {
		if langs[i].count == langs[j].count {
			return langs[i].name < langs[j].name
		}
		return langs[i].count > langs[j].count
	})
	m.langs = langs
	m.langIdx = 0
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

func (m *repositoriesModel) updateLangIdx(reverse bool) {
	n := len(m.langs)
	if reverse {
		m.langIdx = ((m.langIdx-1)%n + n) % n
	} else {
		m.langIdx = (m.langIdx + 1) % n
	}
}

func (m *repositoriesModel) filterItems() {
	if m.langs[m.langIdx].name == "All" {
		m.list.SetItems(m.originalItems)
		return
	}
	items := make([]list.Item, 0)
	for _, i := range m.originalItems {
		if i.(*repositoryItem).langName == m.langs[m.langIdx].name {
			items = append(items, i)
		}
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
		if m.langDialogOpened {
			switch {
			case key.Matches(msg, m.langDialogDelegateKeys.close):
				m.langDialogOpened = false
			case key.Matches(msg, m.langDialogDelegateKeys.next):
				m.list.ResetSelected()
				m.updateLangIdx(false)
				m.filterItems()
			case key.Matches(msg, m.langDialogDelegateKeys.prev):
				m.list.ResetSelected()
				m.updateLangIdx(true)
				m.filterItems()
			}
			return m, nil
		}
		switch {
		case key.Matches(msg, m.delegateKeys.sort):
			m.sortDialogOpened = true
			return m, nil
		case key.Matches(msg, m.delegateKeys.lang):
			m.langDialogOpened = true
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
	if m.langDialogOpened {
		return m.withLangDialogView(ret)
	}
	return ret
}

func (m repositoriesModel) withSortDialogView(base string) string {
	title := repositoriesDialogTitleStyle.Render("Sort")

	body := strings.Join([]string{
		m.sortKeySelectItemView("Stars (Desc)", sortByStarDesc),
		m.sortKeySelectItemView("Stars (Asc)", sortByStarAsc),
		m.sortKeySelectItemView("Last Updated (Desc)", sortByUpdatedDesc),
		m.sortKeySelectItemView("Last Updated (Asc)", sortByUpdatedAsc),
	}, "\n")
	body = repositoriesDialogBodyStyle.Render(body)

	dialog := reposirotiesDialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, body))

	dw, dh := lipgloss.Size(dialog)
	top := (m.height / 2) - (dh / 2)
	left := (m.width / 2) - (dw / 2)
	return kasane.OverlayString(base, dialog, top, left, kasane.WithPadding(m.width))
}

func (m repositoriesModel) sortKeySelectItemView(s string, st sortType) string {
	if m.sortType == st {
		return repositoriesDialogSelectedStyle.Render("> " + s)
	} else {
		return repositoriesDialogNotSelectedStyle.Render("  " + s)
	}
}

func (m repositoriesModel) withLangDialogView(base string) string {
	title := repositoriesDialogTitleStyle.Render("Language")

	ivs := make([]string, len(m.langs))
	for i, l := range m.langs {
		ivs[i] = m.langKeySelectItemView(l)
	}
	body := strings.Join(ivs, "\n")
	body = repositoriesDialogBodyStyle.Render(body)

	dialog := reposirotiesDialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, body))

	dw, dh := lipgloss.Size(dialog)
	top := (m.height / 2) - (dh / 2)
	left := (m.width / 2) - (dw / 2)
	return kasane.OverlayString(base, dialog, top, left, kasane.WithPadding(m.width))
}

func (m repositoriesModel) langKeySelectItemView(lang *repositoeisLang) string {
	if m.langs[m.langIdx].name == lang.name {
		return repositoriesDialogSelectedStyle.Render(fmt.Sprintf("> %s (%d)", lang.name, lang.count))
	} else {
		return repositoriesDialogNotSelectedStyle.Render(fmt.Sprintf("  %s (%d)", lang.name, lang.count))
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
