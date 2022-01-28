package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
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

	profileViewportStyle = lipgloss.NewStyle().
				Padding(1, 0, 0, 0)

	profileSelectedItemColorStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("250")).
					Foreground(lipgloss.Color("56"))
)

type profileSelectableItem int

const (
	profileNotSelectedItem profileSelectableItem = iota
	profileAccountItem
	profileCompanyItem
	profileWebsiteItem
)

type profileKeyMap struct {
	Tab  key.Binding
	Open key.Binding
	Back key.Binding
	Quit key.Binding
}

func (k profileKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Tab,
		k.Open,
		k.Back,
		k.Quit,
	}
}

func (k profileKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Tab,
		},
		{
			k.Open,
		},
		{
			k.Back,
		},
		{
			k.Quit,
		},
	}
}

type profileModel struct {
	client *gh.GitHubClient

	keys         *profileKeyMap
	viewport     viewport.Model
	help         help.Model
	profile      *gh.UserProfile
	spinner      *spinner.Model
	selectedItem profileSelectableItem

	errorMsg      *profileErrorMsg
	loading       bool
	selectedUser  string
	width, height int
}

func newProfileModel(client *gh.GitHubClient, s *spinner.Model) profileModel {
	profileKeys := &profileKeyMap{
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "select item"),
		),
		Open: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "open in browser"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace", "ctrl+h"),
			key.WithHelp("backspace", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
	return profileModel{
		client:       client,
		keys:         profileKeys,
		viewport:     viewport.New(0, 0),
		help:         help.New(),
		spinner:      s,
		selectedItem: profileNotSelectedItem,
	}
}

func (m *profileModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.help.Width = width
	m.viewport.Width = width
	m.viewport.Height = height - 4
}

func (m *profileModel) SetUser(id string) {
	m.selectedUser = id
}

func (m *profileModel) updateProfile(profile *gh.UserProfile) {
	m.profile = profile
	m.updateContent()
}

func (m *profileModel) updateContent() {
	m.viewport.SetContent(m.profieContentsView())
}

func (m *profileModel) selectItem() {
	m.keys.Open.SetEnabled(true)
	m.selectedItem++
	switch m.selectedItem {
	case profileAccountItem:
		// do nothing
	case profileCompanyItem:
		if !isOrganizationLogin(m.profile.Company) {
			m.selectItem()
		}
	case profileWebsiteItem:
		if !isUrl(m.profile.WebsiteUrl) {
			m.selectItem()
		}
	default:
		m.selectedItem = profileNotSelectedItem
		m.keys.Open.SetEnabled(false)
	}
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

func (m profileModel) openInBrowser() tea.Cmd {
	return func() tea.Msg {
		var url string
		switch m.selectedItem {
		case profileAccountItem:
			url = m.profile.Url
		case profileCompanyItem:
			url = organigzationUrlFrom(m.profile.Company)
		case profileWebsiteItem:
			url = m.profile.WebsiteUrl
		default:
			return nil
		}
		if err := openBrowser(url); err != nil {
			return profileErrorMsg{err, "failed to open browser"}
		}
		return nil
	}
}

func (m profileModel) Update(msg tea.Msg) (profileModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Tab):
			m.selectItem()
			m.updateContent()
			return m, nil
		case key.Matches(msg, m.keys.Open):
			return m, m.openInBrowser()
		case key.Matches(msg, m.keys.Back):
			return m, goBackMenuPage
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	case selectProfilePageMsg:
		m.loading = true
		return m, m.loadProfile(msg.id)
	case profileSuccessMsg:
		m.errorMsg = nil
		m.loading = false
		m.selectedItem = profileNotSelectedItem
		m.keys.Open.SetEnabled(false)
		m.updateProfile(msg.profile)
		return m, nil
	case profileErrorMsg:
		m.errorMsg = &msg
		m.loading = false
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m profileModel) View() string {
	if m.loading {
		return loadingView(m.height, m.spinner, m.breadcrumb())
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

	title := titleView(m.breadcrumb())
	ret += title
	height -= cn(title)

	vp := profileViewportStyle.Render(m.viewport.View())
	ret += vp
	height -= cn(vp)

	help := helpStyle.Render(m.help.View(m.keys))
	height -= cn(help)

	ret += strings.Repeat("\n", height)
	ret += help

	return ret
}

func (m profileModel) profieContentsView() string {
	ret := ""
	ret += profileItemNameStyle.Render(m.profile.Name)
	login := "@" + m.profile.Login
	if m.selectedItem == profileAccountItem {
		login = profileSelectedItemColorStyle.Render(login)
	}
	ret += profileItemStyle.Render(login)
	ret += profileItemStyle.Render(m.profile.Bio)
	ret += "\n"
	ret += profileItemStyle.Render(fmt.Sprintf("%d followers - %d following", m.profile.Followers, m.profile.Following))
	company := m.profile.Company
	if m.selectedItem == profileCompanyItem {
		company = profileSelectedItemColorStyle.Render(company)
	}
	ret += profileItemStyle.Render("🏢 " + company)
	ret += profileItemStyle.Render("🌐 " + m.profile.Location)
	websiteUrl := m.profile.WebsiteUrl
	if m.selectedItem == profileWebsiteItem {
		websiteUrl = profileSelectedItemColorStyle.Render(websiteUrl)
	}
	ret += profileItemStyle.Render("🔗 " + websiteUrl)
	return ret
}

func (m profileModel) errorView() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView(m.breadcrumb())
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

func (m profileModel) breadcrumb() []string {
	return []string{m.selectedUser, "Profile"}
}
