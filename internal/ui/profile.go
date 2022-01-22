package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

var (
	profileErrorStyle = lipgloss.NewStyle().
				Padding(2, 0, 0, 2).
				Foreground(lipgloss.Color("161"))

	profileItemStyle = lipgloss.NewStyle().
				Padding(1, 0, 1, 2)

	profileItemNameStyle = profileItemStyle.Copy().
				Bold(true)
)

type profileKeyMap struct {
	Open key.Binding
	Back key.Binding
	Help key.Binding
	Quit key.Binding
}

func (k profileKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Open,
		k.Back,
		k.Help,
		k.Quit,
	}
}

func (k profileKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Open,
		},
		{
			k.Back,
		},
		{
			k.Help,
			k.Quit,
		},
	}
}

type profileModel struct {
	client *gh.GitHubClient

	keys    profileKeyMap
	help    help.Model
	profile *gh.UserProfile
	spinner *spinner.Model

	errorMsg      *profileErrorMsg
	loading       bool
	width, height int
}

func newProfileModel(client *gh.GitHubClient, s *spinner.Model) profileModel {
	profileKeys := profileKeyMap{
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
	return profileModel{
		client:  client,
		keys:    profileKeys,
		help:    help.New(),
		spinner: s,
	}
}

func (m *profileModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.help.Width = width
}

func (m *profileModel) updateProfile(profile *gh.UserProfile) {
	m.profile = profile
}

func (m profileModel) Init() tea.Cmd {
	return nil
}

type profileSuccessMsg struct {
	profile *gh.UserProfile
}

var _ tea.Msg = (*profileSuccessMsg)(nil)

type profileErrorMsg struct {
	e       error
	summary string
}

var _ tea.Msg = (*profileErrorMsg)(nil)

func (m profileModel) loadProfile(id string) tea.Cmd {
	return func() tea.Msg {
		profile, err := m.client.QueryUserProfile(id)
		if err != nil {
			return profileErrorMsg{err, "failed to fetch profile"}
		}
		return profileSuccessMsg{profile}
	}
}

func (m profileModel) openProfilePageInBrowser() tea.Cmd {
	return func() tea.Msg {
		if err := openBrowser(m.profile.Url); err != nil {
			return profileErrorMsg{err, "failed to open browser"}
		}
		return nil
	}
}

func (m profileModel) Update(msg tea.Msg) (profileModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Open):
			return m, m.openProfilePageInBrowser()
		case key.Matches(msg, m.keys.Back):
			return m, goBackMenuPage
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	case selectProfilePageMsg:
		m.loading = true
		return m, m.loadProfile(msg.id)
	case profileSuccessMsg:
		m.errorMsg = nil
		m.loading = false
		m.updateProfile(msg.profile)
		return m, nil
	case profileErrorMsg:
		m.errorMsg = &msg
		m.loading = false
		return m, nil
	}

	return m, nil
}

func (m profileModel) View() string {
	if m.loading {
		return loadingView(m.height, m.spinner)
	}
	if m.errorMsg != nil {
		return m.errorView()
	}
	return m.profieView()
}

func (m profileModel) profieView() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView()
	ret += title
	height -= cn(title)

	name := profileItemNameStyle.Render(m.profile.Name)
	ret += name
	height -= cn(name)

	id := profileItemStyle.Render("@" + m.profile.Login)
	ret += id
	height -= cn(id)

	bio := profileItemStyle.Render(m.profile.Bio)
	ret += bio
	height -= cn(bio)

	ret += "\n"
	height -= 1

	ff := profileItemStyle.Render(fmt.Sprintf("%d followers - %d following", m.profile.Followers, m.profile.Following))
	ret += ff
	height -= cn(ff)

	company := profileItemStyle.Render("🏢 " + m.profile.Company)
	ret += company
	height -= cn(company)

	location := profileItemStyle.Render("🌐 " + m.profile.Location)
	ret += location
	height -= cn(location)

	website := profileItemStyle.Render("🔗 " + m.profile.WebsiteUrl)
	ret += website
	height -= cn(website)

	help := helpStyle.Render(m.help.View(m.keys))
	height -= cn(help)

	ret += strings.Repeat("\n", height)
	ret += help

	return ret
}

func (m profileModel) errorView() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView()
	ret += title
	height -= cn(title)

	errorText := profileErrorStyle.Render("ERROR: " + m.errorMsg.summary)
	ret += errorText
	height -= cn(errorText)

	help := helpStyle.Render(m.help.View(m.keys))
	height -= cn(help)

	ret += strings.Repeat("\n", height)
	ret += help

	return ret
}
