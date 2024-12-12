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

type pullRequestsRepositoryItem struct {
	name        string
	description string
	langName    string
	langColor   string
	prsCount    int
	url         string
}

var _ list.Item = (*pullRequestsRepositoryItem)(nil)

func (i pullRequestsRepositoryItem) FilterValue() string {
	return i.name
}

type pullRequestsRepositoryDelegate struct {
	styles        list.DefaultItemStyles
	shortHelpFunc func() []key.Binding
	fullHelpFunc  func() [][]key.Binding

	normalDescWithoutPadding   lipgloss.Style
	normalDescOnlyPadding      lipgloss.Style
	selectedDescWithoutPadding lipgloss.Style
	selectedDescOnlyPadding    lipgloss.Style
	dimmedDescWithoutPadding   lipgloss.Style
	dimmedDescOnlyPadding      lipgloss.Style
}

var _ list.ItemDelegate = (*pullRequestsRepositoryDelegate)(nil)

func newPullRequestsRepositoryDelegate(delegateKeys pullRequestsRepositoryDelegateKeyMap) pullRequestsRepositoryDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.Copy().Foreground(selectedColor1).BorderForeground(selectedColor2)
	styles.SelectedDesc = styles.SelectedDesc.Copy().Foreground(selectedColor2).BorderForeground(selectedColor2)

	shortHelpFunc := func() []key.Binding {
		return []key.Binding{delegateKeys.sel, delegateKeys.back}
	}
	fullHelpFunc := func() [][]key.Binding {
		return [][]key.Binding{{delegateKeys.open, delegateKeys.sel, delegateKeys.back}}
	}

	normalDescWithoutPadding := styles.NormalDesc.Copy().UnsetPadding()
	normalDescOnlyPadding := lipgloss.NewStyle().Padding(styles.NormalDesc.GetPadding())
	selectedDescWithoutPadding := styles.SelectedDesc.Copy().UnsetPadding().UnsetBorderStyle()
	selectedDescOnlyPadding := lipgloss.NewStyle().Padding(styles.SelectedDesc.GetPadding()).Border(styles.SelectedDesc.GetBorder()).BorderForeground(styles.SelectedDesc.GetBorderLeftForeground())
	dimmedDescWithoutPadding := styles.DimmedDesc.Copy().UnsetPadding()
	dimmedDescOnlyPadding := lipgloss.NewStyle().Padding(styles.DimmedDesc.GetPadding())

	return pullRequestsRepositoryDelegate{
		styles:                     styles,
		shortHelpFunc:              shortHelpFunc,
		fullHelpFunc:               fullHelpFunc,
		normalDescWithoutPadding:   normalDescWithoutPadding,
		normalDescOnlyPadding:      normalDescOnlyPadding,
		selectedDescWithoutPadding: selectedDescWithoutPadding,
		selectedDescOnlyPadding:    selectedDescOnlyPadding,
		dimmedDescWithoutPadding:   dimmedDescWithoutPadding,
		dimmedDescOnlyPadding:      dimmedDescOnlyPadding,
	}
}

func (d pullRequestsRepositoryDelegate) Height() int {
	return 4
}

func (d pullRequestsRepositoryDelegate) Spacing() int {
	return 1
}

func (d pullRequestsRepositoryDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d pullRequestsRepositoryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	matchedRunes := []int{}
	s := &d.styles

	i := item.(*pullRequestsRepositoryItem)
	name := i.name
	desc := i.description
	if desc == "" {
		desc = "-"
	}

	prs := fmt.Sprintf("%d pull request", i.prsCount)
	if i.prsCount > 1 {
		prs += "s"
	}

	// U+25CD
	// U+26AB will be displayed as emoji
	// U+2B24 is too large
	detailsLangColor := "◍ "
	detailsLangColor = lipgloss.NewStyle().Foreground(lipgloss.Color(i.langColor)).Render(detailsLangColor)
	details := fmt.Sprintf("%s     %s", i.langName, prs)

	if m.Width() > 0 {
		textwidth := uint(m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight())
		name = truncate.StringWithTail(name, textwidth, ellipsis)
		desc = truncate.StringWithTail(desc, textwidth, ellipsis)
		// todo: considering max width
	}

	var (
		isSelected  = index == m.Index()
		emptyFilter = m.FilterState() == list.Filtering && m.FilterValue() == ""
		isFiltered  = m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	)

	if isFiltered && index < len(m.VisibleItems()) {
		matchedRunes = m.MatchesForItem(index)
	}

	if emptyFilter {
		name = s.DimmedTitle.Render(name)
		desc = s.DimmedDesc.Render(desc)
		details = d.dimmedDescWithoutPadding.Render(details)
		details = d.dimmedDescOnlyPadding.Render(detailsLangColor + details)
	} else if isSelected && m.FilterState() != list.Filtering {
		if isFiltered {
			unmatched := s.SelectedTitle.Inline(true)
			matched := unmatched.Copy().Inherit(s.FilterMatch)
			name = lipgloss.StyleRunes(name, matchedRunes, matched, unmatched)
		}
		name = s.SelectedTitle.Render(name)
		desc = s.SelectedDesc.Render(desc)
		details = d.selectedDescWithoutPadding.Render(details)
		details = d.selectedDescOnlyPadding.Render(detailsLangColor + details)
	} else {
		if isFiltered {
			unmatched := s.NormalTitle.Inline(true)
			matched := unmatched.Copy().Inherit(s.FilterMatch)
			name = lipgloss.StyleRunes(name, matchedRunes, matched, unmatched)
		}
		name = s.NormalTitle.Render(name)
		desc = s.NormalDesc.Render(desc)
		details = d.normalDescWithoutPadding.Render(details)
		details = d.normalDescOnlyPadding.Render(detailsLangColor + details)
	}

	fmt.Fprintf(w, "%s\n%s\n%s", name, desc, details)
}

func (d pullRequestsRepositoryDelegate) ShortHelp() []key.Binding {
	return d.shortHelpFunc()
}

func (d pullRequestsRepositoryDelegate) FullHelp() [][]key.Binding {
	return d.fullHelpFunc()
}
