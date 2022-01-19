package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type menuModel struct {
	list list.Model

	width, height int
}

func newMenuModel() menuModel {
	items := []list.Item{
		menuItem{
			title:       "Profile",
			description: "Show the user's profile",
		},
		menuItem{
			title:       "PullRequests",
			description: "Show Pull Requests created by the user",
		},
		menuItem{
			title:       "Repositories",
			description: "Show Repositories created by the user",
		},
	}
	// bubbles/list/defaultitem.go
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().Foreground(selectedColor1).BorderForeground(selectedColor2)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Copy().Foreground(selectedColor2).BorderForeground(selectedColor2)
	l := list.New(items, delegate, 0, 0)
	l.Title = appTitle
	l.Styles.Title = titleStyle
	return menuModel{
		list: l,
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

func (m *menuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (menuModel, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	list, cmd := m.list.Update(msg)
	m.list = list
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m menuModel) View() string {
	return m.list.View()
}
