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
}

var _ list.Item = (*repositoryItem)(nil)

func (i repositoryItem) FilterValue() string {
	return i.title
}

type repositoryDelegate struct {
	styles        list.DefaultItemStyles
	shortHelpFunc func() []key.Binding
	fullHelpFunc  func() [][]key.Binding

	normalDescWithoutPadding   lipgloss.Style
	normalDescOnlyPadding      lipgloss.Style
	selectedDescWithoutPadding lipgloss.Style
	selectedDescOnlyPadding    lipgloss.Style
}

var _ list.ItemDelegate = (*repositoryDelegate)(nil)

func NewRepositoryDelegate(delegateKeys repositoriesDelegateKeyMap) repositoryDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.Copy().Foreground(selectedColor1).BorderForeground(selectedColor2)
	styles.SelectedDesc = styles.SelectedDesc.Copy().Foreground(selectedColor2).BorderForeground(selectedColor2)

	shortHelpFunc := func() []key.Binding {
		return []key.Binding{delegateKeys.open, delegateKeys.back}
	}
	fullHelpFunc := func() [][]key.Binding {
		return [][]key.Binding{{delegateKeys.open, delegateKeys.back}}
	}

	normalDescWithoutPadding := styles.NormalDesc.Copy().UnsetPadding()
	normalDescOnlyPadding := lipgloss.NewStyle().Padding(styles.NormalDesc.GetPadding())
	selectedDescWithoutPadding := styles.SelectedDesc.Copy().UnsetPadding().UnsetBorderStyle()
	selectedDescOnlyPadding := lipgloss.NewStyle().Padding(styles.SelectedDesc.GetPadding()).Border(styles.SelectedDesc.GetBorder())

	return repositoryDelegate{
		styles:                     styles,
		shortHelpFunc:              shortHelpFunc,
		fullHelpFunc:               fullHelpFunc,
		normalDescWithoutPadding:   normalDescWithoutPadding,
		normalDescOnlyPadding:      normalDescOnlyPadding,
		selectedDescWithoutPadding: selectedDescWithoutPadding,
		selectedDescOnlyPadding:    selectedDescOnlyPadding,
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
	s := &d.styles

	i := item.(*repositoryItem)
	title := i.title
	desc := i.description
	if desc == "" {
		desc = "-"
	}

	// U+25CD
	// U+26AB will be displayed as emoji
	// U+2B24 is too large
	detailsLangColor := "◍ "
	detailsLangColor = lipgloss.NewStyle().Foreground(lipgloss.Color(i.langColor)).Render(detailsLangColor)
	license := i.license
	if license == "" {
		license = "-"
	}
	details := fmt.Sprintf("%s   ⚖ %s   Updated %s", i.langName, license, i.updated)

	counts := fmt.Sprintf("⭐ %d   🍴 %d   👀 %d", i.stars, i.forks, i.watchers)

	if m.Width() > 0 {
		textwidth := uint(m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight())
		title = truncate.StringWithTail(title, textwidth, ellipsis)
		desc = truncate.StringWithTail(desc, textwidth, ellipsis)
		// todo: considering max width
	}

	if index == m.Index() {
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)
		details = d.selectedDescWithoutPadding.Render(details)
		details = d.selectedDescOnlyPadding.Render(detailsLangColor + details)
		counts = s.SelectedDesc.Render(counts)
	} else {
		title = s.NormalTitle.Render(title)
		desc = s.NormalDesc.Render(desc)
		details = d.normalDescWithoutPadding.Render(details)
		details = d.normalDescOnlyPadding.Render(detailsLangColor + details)
		counts = s.NormalDesc.Render(counts)
	}

	fmt.Fprintf(w, "%s\n%s\n%s\n%s", title, desc, counts, details)
}

func (d repositoryDelegate) ShortHelp() []key.Binding {
	return d.shortHelpFunc()
}

func (d repositoryDelegate) FullHelp() [][]key.Binding {
	return d.fullHelpFunc()
}
