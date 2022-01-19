package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lusingander/ghcv-cli/internal/gh"
)

var (
	inputLabelStyle = lipgloss.NewStyle().
			Padding(1, 0, 1, 2)

	inputUserStyle = lipgloss.NewStyle().
			Padding(1, 0, 0, 2)

	inputSpinnerStyle = lipgloss.NewStyle().
				Padding(2, 0, 0, 2)

	inputErrorStyle = lipgloss.NewStyle().
			Padding(2, 0, 0, 2).
			Foreground(lipgloss.Color("161"))
)

type userSelectModel struct {
	client *gh.GitHubClient

	keys    userSelectKeyMap
	input   textinput.Model
	help    help.Model
	spinner spinner.Model

	errorMsg      *userSelectErrorMsg
	loading       bool
	width, height int
}

type userSelectKeyMap struct {
	Enter key.Binding
	Help  key.Binding
	Quit  key.Binding
}

func (k userSelectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Enter,
		k.Help,
		k.Quit,
	}
}

func (k userSelectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Enter,
		},
		{
			k.Help,
			k.Quit,
		},
	}
}

func newUserSelectModel(client *gh.GitHubClient) userSelectModel {
	userSelectKeys := userSelectKeyMap{
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
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

	inputModel := textinput.New()
	inputModel.Placeholder = "GitHub ID"
	inputModel.Focus()

	s := spinner.New()
	s.Spinner = spinner.Moon

	return userSelectModel{
		client:  client,
		keys:    userSelectKeys,
		input:   inputModel,
		help:    help.New(),
		spinner: s,
	}
}

func (m *userSelectModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.help.Width = width
}

func (m userSelectModel) Init() tea.Cmd {
	return nil
}

type userSelectSuccessMsg struct {
	id string
}

var _ tea.Msg = (*userSelectSuccessMsg)(nil)

type userSelectErrorMsg struct {
	e      error
	detail string
}

var _ tea.Msg = (*userSelectErrorMsg)(nil)

func (m userSelectModel) checkUser() tea.Cmd {
	id := strings.TrimSpace(m.input.Value())
	if id == "" {
		return nil
	}
	return func() tea.Msg {
		_, err := m.client.QueryUserProfile(id)
		if err != nil {
			return userSelectErrorMsg{err, "user not found"}
		}
		return userSelectSuccessMsg{id}
	}
}

func (m userSelectModel) Update(msg tea.Msg) (userSelectModel, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			cmd := m.checkUser()
			if cmd == nil {
				return m, nil
			}
			m.input.Blur()
			m.errorMsg = nil
			m.loading = true
			return m, cmd
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	case userSelectSuccessMsg:
		m.errorMsg = nil
		m.loading = false
		return m, userSelected(msg.id)
	case userSelectErrorMsg:
		m.errorMsg = &msg
		m.loading = false
		m.input.Focus()
		return m, nil
	}

	input, iCmd := m.input.Update(msg)
	m.input = input
	cmds = append(cmds, iCmd)

	sp, sCmd := m.spinner.Update(msg)
	m.spinner = sp
	cmds = append(cmds, sCmd)

	return m, tea.Batch(cmds...)
}

func (m userSelectModel) View() string {
	if m.height <= 0 {
		return ""
	}

	ret := ""
	height := m.height - 1

	title := titleView()
	ret += title
	height -= cn(title)

	label := inputLabelStyle.Render("Enter GitHub User ID")
	ret += label
	height -= cn(label)

	input := inputUserStyle.Render(m.input.View())
	ret += input
	height -= cn(input)

	if m.loading {
		sp := inputSpinnerStyle.Render(m.spinner.View() + " Loading...")
		ret += sp
		height -= cn(sp)
	}

	if m.errorMsg != nil {
		errorText := inputErrorStyle.Render("ERROR: " + m.errorMsg.detail)
		ret += errorText
		height -= cn(errorText)
	}

	help := helpStyle.Render(m.help.View(m.keys))
	height -= cn(help)

	ret += strings.Repeat("\n", height)
	ret += help

	return ret
}

func cn(view string) int {
	return strings.Count(view, "\n")
}
