package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

var (
	repositoriesErrorStyle = lipgloss.NewStyle().
		Padding(2, 0, 0, 2).
		Foreground(lipgloss.Color("161"))
)

type repositoriesModel struct {
	client *gh.GitHubClient

	list    list.Model
	spinner *spinner.Model

	delegateKeys repositoriesDelegateKeyMap

	errorMsg      *repositoriesErrorMsg
	loading       bool
	selectedUser  string
	width, height int
}

type repositoriesDelegateKeyMap struct {
	open key.Binding
	back key.Binding
	quit key.Binding
}

func newRepositoriesDelegateKeyMap() repositoriesDelegateKeyMap {
	return repositoriesDelegateKeyMap{
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

func newRepositoriesModel(client *gh.GitHubClient, s *spinner.Model) repositoriesModel {
	delegateKeys := newRepositoriesDelegateKeyMap()
	delegate := NewRepositoryDelegate(delegateKeys)

	l := list.New(nil, delegate, 0, 0)
	l.KeyMap.Quit = delegateKeys.quit
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)

	return repositoriesModel{
		client:       client,
		list:         l,
		spinner:      s,
		delegateKeys: delegateKeys,
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
		}
		items[i] = item
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
		switch {
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
	return titleView(m.breadcrumb()) + listView(m.list)
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
