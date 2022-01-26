package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

var (
	statusStyleBase = lipgloss.NewStyle().
			Underline(true)

	statusOpenStyle = statusStyleBase.Copy().
			Bold(true).
			Foreground(lipgloss.Color("34"))

	statusMergedStyle = statusStyleBase.Copy().
				Bold(true).
				Foreground(lipgloss.Color("98"))

	statusClosedStyle = statusStyleBase.Copy().
				Bold(true).
				Foreground(lipgloss.Color("203"))

	additionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34"))

	deletionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))
)

type pullRequestsListItem struct {
	title     string
	status    string
	number    int
	additions int
	deletions int
	comments  int
	created   string
	closed    string
	url       string
}

func (i pullRequestsListItem) styledTitle(selected bool) string {
	var title, status string
	if selected {
		title = listSelectedTitleColorStyle.Render(i.title)
	} else {
		title = listNormalTitleColorStyle.Render(i.title)
	}
	switch i.status {
	case "OPEN":
		status = statusOpenStyle.Render(i.status)
	case "MERGED":
		status = statusMergedStyle.Render(i.status)
	case "CLOSED":
		status = statusClosedStyle.Render(i.status)
	}
	return fmt.Sprintf("%s  %s", status, title)
}

func (i pullRequestsListItem) styledDesc(selected bool) string {
	num := i.styledNumber(selected)
	upd := i.styledUpdate(selected)
	mods := i.styledModifications()
	return fmt.Sprintf("%s  %s  %s", num, upd, mods)
}

func (i pullRequestsListItem) styledNumber(selected bool) string {
	s := fmt.Sprintf("#%d", i.number)
	if selected {
		return listSelectedDescColorStyle.Render(s)
	}
	return listNormalDescColorStyle.Render(s)
}

func (i pullRequestsListItem) styledUpdate(selected bool) string {
	var upd, st string
	switch i.status {
	case "OPEN":
		upd = i.created
		st = "opened"
	case "MERGED":
		upd = i.closed
		st = "merged"
	case "CLOSED":
		upd = i.closed
		st = "closed"
	}
	s := fmt.Sprintf("%s %s", st, upd)
	if selected {
		return listSelectedDescColorStyle.Render(s)
	}
	return listNormalDescColorStyle.Render(s)
}

func (i pullRequestsListItem) styledModifications() string {
	s := ""
	if i.additions > 0 {
		s += additionsStyle.Render(fmt.Sprintf("+%d", i.additions))
	}
	if i.deletions > 0 {
		s += deletionsStyle.Render(fmt.Sprintf("-%d", i.deletions))
	}
	return s
}

var _ list.Item = (*pullRequestsListItem)(nil)

func (i pullRequestsListItem) FilterValue() string {
	return i.title
}

type pullRequestsListDelegate struct {
	shortHelpFunc func() []key.Binding
	fullHelpFunc  func() [][]key.Binding
}

var _ list.ItemDelegate = (*pullRequestsListDelegate)(nil)

func newPullRequestsListDelegate(delegateKeys pullRequestsListDelegateKeyMap) pullRequestsListDelegate {
	shortHelpFunc := func() []key.Binding {
		return []key.Binding{delegateKeys.back}
	}
	fullHelpFunc := func() [][]key.Binding {
		return [][]key.Binding{{delegateKeys.open, delegateKeys.back}}
	}
	return pullRequestsListDelegate{
		shortHelpFunc: shortHelpFunc,
		fullHelpFunc:  fullHelpFunc,
	}
}

func (d pullRequestsListDelegate) Height() int {
	return 2
}

func (d pullRequestsListDelegate) Spacing() int {
	return 1
}

func (d pullRequestsListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d pullRequestsListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	selected := index == m.Index()

	i := item.(pullRequestsListItem)
	title := i.styledTitle(selected)
	desc := i.styledDesc(selected)

	if m.Width() > 0 {
		textwidth := uint(m.Width() - listNormalTitleStyle.GetPaddingLeft() - listNormalTitleStyle.GetPaddingRight())
		title = truncate.StringWithTail(title, textwidth, ellipsis)
		// desc = truncate.StringWithTail(desc, textwidth, ellipsis)
		// todo: considering max width
	}

	if selected {
		title = listSelectedItemStyle.Render(title)
		desc = listSelectedItemStyle.Render(desc)
	} else {
		title = listNormalItemStyle.Render(title)
		desc = listNormalItemStyle.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

func (d pullRequestsListDelegate) ShortHelp() []key.Binding {
	return d.shortHelpFunc()
}

func (d pullRequestsListDelegate) FullHelp() [][]key.Binding {
	return d.fullHelpFunc()
}
