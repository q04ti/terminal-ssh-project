package ui

import (
	"fmt"
	"strings"

	"afterdark/internal/chat"
	"afterdark/internal/security"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LobbyModel handles the live multiplayer chat room.
type LobbyModel struct {
	viewport   viewport.Model
	textInput  textinput.Model
	styles     Styles
	messages   []chat.Message
	users      []chat.UserInfo
	muted      bool
	errMessage string
	width      int
	height     int
	ready      bool
}

// NewLobbyModel initializes the lobby UI.
func NewLobbyModel(styles Styles, initialHistory []chat.Message) LobbyModel {
	ti := textinput.New()
	ti.Placeholder = "type a message... (Esc to return to hub)"
	ti.Focus()
	ti.CharLimit = 280
	ti.Width = 60

	return LobbyModel{
		textInput: ti,
		styles:    styles,
		messages:  initialHistory,
	}
}

// SetSize updates dimensions of viewport and input.
func (m *LobbyModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	vpHeight := height - 10
	if vpHeight < 5 {
		vpHeight = 5
	}
	vpWidth := width - 26 // leave room for online users sidebar
	if vpWidth < 30 {
		vpWidth = width - 4
	}

	if !m.ready {
		m.viewport = viewport.New(vpWidth, vpHeight)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		m.ready = true
	} else {
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
		m.viewport.SetContent(m.renderMessages())
	}
	m.textInput.Width = vpWidth - 10
}

// AddMessage appends a message and scrolls.
func (m *LobbyModel) AddMessage(msg chat.Message) {
	m.messages = append(m.messages, msg)
	if len(m.messages) > 150 {
		m.messages = m.messages[len(m.messages)-150:]
	}
	if m.ready {
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	}
}

// SetUsers updates online visitor list.
func (m *LobbyModel) SetUsers(users []chat.UserInfo) {
	m.users = users
}

// Update handles input events in the lobby.
func (m *LobbyModel) Update(msg tea.Msg) (tea.Cmd, *string, bool) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.errMessage = ""
		switch msg.Type {
		case tea.KeyEnter:
			val := strings.TrimSpace(m.textInput.Value())
			if val == "" {
				return nil, nil, false
			}
			clean := security.SanitizeText(val, 280, false)
			if clean == "" {
				m.errMessage = "Message cannot be empty."
				return nil, nil, false
			}
			m.textInput.Reset()
			return nil, &clean, false

		case tea.KeyEsc:
			return nil, nil, true // return to hub

		case tea.KeyPgUp, tea.KeyPgDown:
			m.viewport, cmd = m.viewport.Update(msg)
			return cmd, nil, false
		}

		if msg.String() == "ctrl+m" {
			m.muted = !m.muted
			return nil, nil, false
		}
	}

	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)
	return tiCmd, nil, false
}

func (m *LobbyModel) renderMessages() string {
	if len(m.messages) == 0 {
		return m.styles.Muted.Render("The air is quiet. Speak into the wires...")
	}

	var sb strings.Builder
	for _, msg := range m.messages {
		ts := msg.Timestamp.Format("15:04")
		switch msg.Type {
		case chat.MsgSystem, chat.MsgJoin, chat.MsgLeave:
			line := fmt.Sprintf("[%s] * %s\n", ts, msg.Content)
			sb.WriteString(m.styles.Muted.Render(line))
		default:
			prefix := fmt.Sprintf("[%s] %s: ", ts, msg.Author)
			authorStyle := m.styles.Highlight
			if msg.ColorHex != "" {
				authorStyle = authorStyle.Foreground(lipgloss.Color(msg.ColorHex))
			}
			line := authorStyle.Render(prefix) + m.styles.MenuItem.Render(msg.Content) + "\n"
			sb.WriteString(line)
		}
	}
	return sb.String()
}

// View renders the lobby screen.
func (m LobbyModel) View() string {
	var b strings.Builder

	// Header
	header := m.styles.Title.Render("AFTERDARK") + m.styles.Muted.Render(" / ") + m.styles.Prompt.Render("THE LOBBY")
	muteStatus := ""
	if m.muted {
		muteStatus = m.styles.Error.Render(" [MUTED - Ctrl+M to unmute]")
	} else {
		muteStatus = m.styles.Muted.Render(" [Ctrl+M to mute]")
	}
	b.WriteString(header + muteStatus + "\n")
	b.WriteString(m.styles.Muted.Render(strings.Repeat("─", max(20, m.width-4))) + "\n")

	// Render messages viewport and user sidebar
	sidebarWidth := 22
	userListStr := m.renderUserList(sidebarWidth)

	vpBox := m.viewport.View()
	content := ""
	if m.width >= 70 {
		content = lipgloss.JoinHorizontal(lipgloss.Top, vpBox, "  ", userListStr)
	} else {
		content = vpBox
	}
	b.WriteString(content + "\n")

	b.WriteString(m.styles.Muted.Render(strings.Repeat("─", max(20, m.width-4))) + "\n")

	if m.errMessage != "" {
		b.WriteString(m.styles.Error.Render("! "+m.errMessage) + "\n")
	}

	// Input bar
	prompt := m.styles.Prompt.Render("Message: ") + m.textInput.View()
	b.WriteString(prompt + "\n")
	b.WriteString(m.styles.Muted.Render("[Enter] Send  •  [Esc] Hub  •  [PgUp/PgDn] Scroll  •  [Ctrl+M] Mute"))

	return b.String()
}

func (m *LobbyModel) renderUserList(width int) string {
	var sb strings.Builder
	sb.WriteString(m.styles.Highlight.Render(fmt.Sprintf("ONLINE: %02d", len(m.users))) + "\n")
	sb.WriteString(m.styles.Muted.Render(strings.Repeat("·", width-2)) + "\n")

	displayCount := 0
	for _, u := range m.users {
		if displayCount >= 12 {
			sb.WriteString(m.styles.Muted.Render(fmt.Sprintf("... +%d more", len(m.users)-displayCount)) + "\n")
			break
		}
		nick := u.Nickname
		if len(nick) > 12 {
			nick = nick[:12]
		}
		room := u.Room
		if room == "" {
			room = "Hub"
		}
		sb.WriteString(m.styles.MenuItem.Render("• "+nick) + " " + m.styles.Muted.Render("("+room+")") + "\n")
		displayCount++
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(m.styles.Theme.Border).
		PaddingLeft(1).
		Width(width)

	return boxStyle.Render(sb.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
