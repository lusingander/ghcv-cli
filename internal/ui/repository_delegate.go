package ui

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

const (
	ellipsis = "…"
)

type repositoryItem struct {
	title       string
	description string
	langName    string
	langColor   string
	license     string
	updated     string
	stars       int
	forks       int
	watchers    int
	url         string
	pushedAt    time.Time
}

func (i repositoryItem) titleStr() string {
	return i.title
}

func (i repositoryItem) descStr() string {
	if i.description == "" {
		return "-"
	}
	return i.description
}

func (i repositoryItem) styledLangColor() string {
	// U+25CD
	// U+26AB will be displayed as emoji
	// U+2B24 is too large
	s := "◍ "
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(i.langColor))
	return style.Render(s)
}

func (i repositoryItem) detailsStr() string {
	license := i.license
	if license == "" {
		license = "-"
	}
	return fmt.Sprintf("%s   ⚖ %s   Updated %s", i.langName, license, i.updated)
}

func (i repositoryItem) countsStr() string {
	return fmt.Sprintf("Star: %d / Fork: %d / Watch: %d", i.stars, i.forks, i.watchers)
}

var _ list.Item = (*repositoryItem)(nil)

func (i repositoryItem) FilterValue() string {
	return i.title
}

type repositoryDelegate struct {
	shortHelpFunc func() []key.Binding
	fullHelpFunc  func() [][]key.Binding
}

var _ list.ItemDelegate = (*repositoryDelegate)(nil)

func NewRepositoryDelegate(delegateKeys repositoriesDelegateKeyMap) repositoryDelegate {

	shortHelpFunc := func() []key.Binding {
		return []key.Binding{delegateKeys.open, delegateKeys.back}
	}
	fullHelpFunc := func() [][]key.Binding {
		return [][]key.Binding{{delegateKeys.sort, delegateKeys.lang, delegateKeys.open, delegateKeys.back}}
	}

	return repositoryDelegate{
		shortHelpFunc: shortHelpFunc,
		fullHelpFunc:  fullHelpFunc,
	}
}

func (d repositoryDelegate) Height() int {
	return 4
}

func (d repositoryDelegate) Spacing() int {
	return 1
}

func (d repositoryDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d repositoryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i := item.(*repositoryItem)
	title := i.titleStr()
	desc := i.descStr()
	detailsLangColor := i.styledLangColor()
	details := i.detailsStr()
	counts := i.countsStr()

	if m.Width() > 0 {
		textwidth := uint(m.Width() - listNormalTitleStyle.GetPaddingLeft() - listNormalTitleStyle.GetPaddingRight())
		title = truncate.StringWithTail(title, textwidth, ellipsis)
		desc = truncate.StringWithTail(desc, textwidth, ellipsis)
		// todo: considering max width
	}

	if index == m.Index() {
		title = listSelectedTitleStyle.Render(title)
		desc = listSelectedDescStyle.Render(desc)
		counts = listSelectedDescStyle.Render(counts)
		details = listSelectedDescColorStyle.Render(details)
		details = listSelectedItemStyle.Render(detailsLangColor + details)
	} else {
		title = listNormalTitleStyle.Render(title)
		desc = listNormalDescStyle.Render(desc)
		counts = listNormalDescStyle.Render(counts)
		details = listNormalDescColorStyle.Render(details)
		details = listNormalItemStyle.Render(detailsLangColor + details)
	}

	fmt.Fprintf(w, "%s\n%s\n%s\n%s", title, desc, counts, details)
}

func (d repositoryDelegate) ShortHelp() []key.Binding {
	return d.shortHelpFunc()
}

func (d repositoryDelegate) FullHelp() [][]key.Binding {
	return d.fullHelpFunc()
}
