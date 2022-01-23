package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	helpTitleAbout   = "About"
	helpTitleCredits = "Credits"
)

type helpModel struct {
	list list.Model

	delegateKeys helpDelegateKeyMap

	width, height int
}

type helpItem struct {
	title string
	desc  string
}

var _ list.DefaultItem = (*helpItem)(nil)

func (i helpItem) Title() string {
	return i.title
}

func (i helpItem) Description() string {
	return i.desc
}

func (i helpItem) FilterValue() string {
	return i.title
}

func newHelpModel() helpModel {
	items := []list.Item{
		helpItem{
			title: helpTitleAbout,
			desc:  "Show about this application",
		},
		helpItem{
			title: helpTitleCredits,
			desc:  "Show license information for this application",
		},
	}

	delegate := list.NewDefaultDelegate()

	delegateKeys := newHelpDelegateKeyMap()
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
	return helpModel{
		list:         l,
		delegateKeys: delegateKeys,
	}
}

type helpDelegateKeyMap struct {
	back key.Binding
	sel  key.Binding
}

func newHelpDelegateKeyMap() helpDelegateKeyMap {
	return helpDelegateKeyMap{
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

func (m *helpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-2)
}

func (m helpModel) Init() tea.Cmd {
	return nil
}

func (m helpModel) Update(msg tea.Msg) (helpModel, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.delegateKeys.sel):
			switch m.list.SelectedItem().(helpItem).Title() {
			case helpTitleAbout:
				return m, selectAboutPage
			case helpTitleCredits:
				// todo
			}
		case key.Matches(msg, m.delegateKeys.back):
			return m, goBackMenuPage
		}
	case selectHelpPageMsg:
		m.list.ResetSelected()
	}

	list, cmd := m.list.Update(msg)
	m.list = list
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m helpModel) View() string {
	return titleView(m.breadcrumb()) + listView(m.list)
}

func (m helpModel) breadcrumb() []string {
	return []string{"Help"}
}
