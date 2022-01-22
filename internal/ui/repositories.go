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
	width, height int
}

// todo: fix
type repositoryItem struct {
	title       string
	description string
}

var _ list.DefaultItem = (*repositoryItem)(nil)

func (i repositoryItem) Title() string {
	return i.title
}

func (i repositoryItem) Description() string {
	if i.description == "" {
		return "-"
	}
	return i.description
}

func (i repositoryItem) FilterValue() string {
	return i.title
}

type repositoriesDelegateKeyMap struct {
	back key.Binding
}

func newRepositoriesDelegateKeyMap() repositoriesDelegateKeyMap {
	return repositoriesDelegateKeyMap{
		back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
	}
}

func newRepositoriesModel(client *gh.GitHubClient, s *spinner.Model) repositoriesModel {
	var items []list.Item
	delegate := list.NewDefaultDelegate()

	delegateKeys := newRepositoriesDelegateKeyMap()
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
	m.list.SetSize(width, height)
}

func (m *repositoriesModel) updateItems(repos *gh.UserRepositories) {
	items := make([]list.Item, len(repos.Repositories))
	for i, repo := range repos.Repositories {
		item := &repositoryItem{
			title:       repo.Name,
			description: repo.Description,
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

func (m repositoriesModel) Update(msg tea.Msg) (repositoriesModel, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
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
		return loadingView(m.height, m.spinner)
	}
	if m.errorMsg != nil {
		return m.errorView()
	}
	return m.list.View()
}

func (m repositoriesModel) errorView() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView()
	ret += title
	height -= cn(title)

	errorText := repositoriesErrorStyle.Render("ERROR: " + m.errorMsg.summary)
	ret += errorText
	height -= cn(errorText)

	return ret
}
