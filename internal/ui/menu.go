package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	menuTitleProfile      = "Profile"
	menuTitlePullRequests = "Pull Requests"
	menuTitleRepositories = "Repositories"
	menuTitleHelp         = "Help"
)

type menuModel struct {
	list list.Model

	delegateKeys menuDelegateKeyMap

	selectedUser  string
	width, height int
}

func newMenuModel() menuModel {
	items := []list.Item{
		menuItem{
			title:       menuTitleProfile,
			description: "Show the user's profile",
		},
		menuItem{
			title:       menuTitlePullRequests,
			description: "Show Pull Requests created by the user",
		},
		menuItem{
			title:       menuTitleRepositories,
			description: "Show Repositories created by the user",
		},
		menuItem{
			title:       menuTitleHelp,
			description: "Show help menus",
		},
	}

	delegate := list.NewDefaultDelegate()

	delegateKeys := newMenuDelegateKeyMap()
	delegate.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{delegateKeys.sel, delegateKeys.back}
	}
	delegate.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{{delegateKeys.sel, delegateKeys.back}}
	}

	// bubbles/list/defaultitem.go
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().Foreground(selectedColor1).BorderForeground(selectedColor2)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().Foreground(selectedColor2).BorderForeground(selectedColor2)
	l := list.New(items, delegate, 0, 0)
	l.KeyMap.Quit = key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c", "quit"),
	)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	return menuModel{
		list:         l,
		delegateKeys: delegateKeys,
	}
}

type menuItem struct {
	title       string
	description string
}

var _ list.DefaultItem = (*menuItem)(nil)

func (i menuItem) Title() string {
	return i.title
}

func (i menuItem) Description() string {
	return i.description
}

func (i menuItem) FilterValue() string {
	return i.title
}

type menuDelegateKeyMap struct {
	back key.Binding
	sel  key.Binding
}

func newMenuDelegateKeyMap() menuDelegateKeyMap {
	return menuDelegateKeyMap{
		back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
		sel: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
	}
}

func (m *menuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-2)
}

func (m *menuModel) SetUser(id string) {
	m.selectedUser = id
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (menuModel, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.delegateKeys.sel):
			switch m.list.SelectedItem().(menuItem).Title() {
			case menuTitleProfile:
				return m, selectProfilePage(m.selectedUser)
			case menuTitleRepositories:
				return m, selectRepositoriesPage(m.selectedUser)
			case menuTitlePullRequests:
				return m, selectPullRequestsPage(m.selectedUser)
			case menuTitleHelp:
				return m, selectHelpPage
			}
		case key.Matches(msg, m.delegateKeys.back):
			return m, goBackUserSelectPage
		}
	case userSelectMsg:
		m.list.ResetSelected()
	}

	list, cmd := m.list.Update(msg)
	m.list = list
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m menuModel) View() string {
	return titleView(m.breadcrumb()) + listView(m.list)
}

func (m menuModel) breadcrumb() []string {
	return []string{m.selectedUser}
}
