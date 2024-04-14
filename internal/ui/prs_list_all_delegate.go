package ui

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/truncate"
)

type pullRequestsListAllItem struct {
	owner      string
	repository string
	createdAt  time.Time
	closedAt   time.Time
	pullRequestsListItem
}

func (i pullRequestsListAllItem) styledRepo(selected bool) string {
	name := fmt.Sprintf("%s/%s", i.owner, i.repository)
	if selected {
		name = listSelectedTitleColorStyle.Render(name)
	} else {
		name = listNormalTitleColorStyle.Render(name)
	}
	return name
}

var _ list.Item = (*pullRequestsListAllItem)(nil)

func (i pullRequestsListAllItem) FilterValue() string {
	return i.title
}

type pullRequestsListAllDelegate struct {
	shortHelpFunc func() []key.Binding
	fullHelpFunc  func() [][]key.Binding
}

var _ list.ItemDelegate = (*pullRequestsListAllDelegate)(nil)

func newPullRequestsListAllDelegate(delegateKeys pullRequestsListAllDelegateKeyMap) pullRequestsListAllDelegate {
	shortHelpFunc := func() []key.Binding {
		return []key.Binding{delegateKeys.back, delegateKeys.tog}
	}
	fullHelpFunc := func() [][]key.Binding {
		return [][]key.Binding{{delegateKeys.stat, delegateKeys.open, delegateKeys.back, delegateKeys.tog}}
	}
	return pullRequestsListAllDelegate{
		shortHelpFunc: shortHelpFunc,
		fullHelpFunc:  fullHelpFunc,
	}
}

func (d pullRequestsListAllDelegate) Height() int {
	return 3
}

func (d pullRequestsListAllDelegate) Spacing() int {
	return 1
}

func (d pullRequestsListAllDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d pullRequestsListAllDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	selected := index == m.Index()

	i := item.(pullRequestsListAllItem)
	repo := i.styledRepo(selected)
	title := i.styledTitle(selected)
	desc := i.styledDesc(selected)

	if m.Width() > 0 {
		textwidth := uint(m.Width() - listNormalTitleStyle.GetPaddingLeft() - listNormalTitleStyle.GetPaddingRight())
		title = truncate.StringWithTail(title, textwidth, ellipsis)
		// todo: considering max width
	}

	if selected {
		repo = listSelectedItemStyle.Render(repo)
		title = listSelectedItemStyle.Render(title)
		desc = listSelectedItemStyle.Render(desc)
	} else {
		repo = listNormalItemStyle.Render(repo)
		title = listNormalItemStyle.Render(title)
		desc = listNormalItemStyle.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s\n%s", repo, title, desc)
}

func (d pullRequestsListAllDelegate) ShortHelp() []key.Binding {
	return d.shortHelpFunc()
}

func (d pullRequestsListAllDelegate) FullHelp() [][]key.Binding {
	return d.fullHelpFunc()
}
