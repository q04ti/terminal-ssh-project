package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HubMenuOption represents a menu item.
type HubMenuOption struct {
	Key         string
	Title       string
	Description string
	RoomKey     string
}

// HubViewModel renders the main terminal hub.
type HubViewModel struct {
	styles      Styles
	options     []HubMenuOption
	cursor      int
	onlineUsers int
	serverName  string
	width       int
	height      int
}

// NewHubViewModel initializes the hub view.
func NewHubViewModel(styles Styles, serverName string) HubViewModel {
	options := []HubMenuOption{
		{"1", "THE LOBBY", "Join real-time conversation with active explorers", "Lobby"},
		{"2", "THE ARCHIVE", "Read and preserve anonymous transmissions across time", "Archive"},
		{"3", "THE WALL", "Public graffiti board for persistent visitor tags", "Wall"},
		{"4", "THE VOID", "Quiet room to release thoughts into absolute silence", "Void"},
		{"5", "THE NETWORK", "Live telemetry, node topology, and activity metrics", "Network"},
		{"6", "THE PROFILE", "Personalize handle, terminal theme, view badges", "Profile"},
		{"7", "THE SECRETS", "Terminal command interface for hidden protocols", "Secrets"},
		{"Q", "DISCONNECT", "Sever SSH connection gracefully", "Disconnect"},
	}

	return HubViewModel{
		styles:     styles,
		options:    options,
		cursor:     0,
		serverName: serverName,
	}
}

// SetOnlineUsers updates active user count.
func (m *HubViewModel) SetOnlineUsers(count int) {
	m.onlineUsers = count
}

// SetSize updates dimensions.
func (m *HubViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles arrow keys and number shortcuts.
func (m *HubViewModel) Update(msg tea.Msg) (string, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.options) - 1
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "enter":
			return m.options[m.cursor].RoomKey, true
		case "1":
			return "Lobby", true
		case "2":
			return "Archive", true
		case "3":
			return "Wall", true
		case "4":
			return "Void", true
		case "5":
			return "Network", true
		case "6":
			return "Profile", true
		case "7":
			return "Secrets", true
		case "q", "Q":
			return "Disconnect", true
		}
	}
	return "", false
}

// View renders the hub screen.
func (m HubViewModel) View() string {
	var b strings.Builder

	// Top Header Bar
	headerLeft := m.styles.Title.Render(m.serverName)
	headerRight := m.styles.Highlight.Render(fmt.Sprintf("● ONLINE: %02d", m.onlineUsers))
	spaceWidth := max(2, m.width-lipgloss.Width(headerLeft)-lipgloss.Width(headerRight)-4)
	header := headerLeft + strings.Repeat(" ", spaceWidth) + headerRight

	b.WriteString(header + "\n")
	b.WriteString(m.styles.Muted.Render(strings.Repeat("━", max(20, m.width-4))) + "\n")

	// Logo: choose Big or Compact based on width
	if m.width >= 90 {
		b.WriteString(m.styles.Title.Render(ASCIILogoBig) + "\n\n")
	} else {
		b.WriteString(m.styles.Title.Render(ASCIILogoCompact) + "\n\n")
	}

	b.WriteString(m.styles.Subtitle.Render("   An internet hidden inside SSH. Forgotten wires and quiet corners.") + "\n\n")

	// Menu options
	for i, opt := range m.options {
		isSelected := i == m.cursor
		prefix := "  "
		keyStr := m.styles.MenuKey.Render("[" + opt.Key + "]")
		titleStr := m.styles.MenuItem.Render(opt.Title)

		if isSelected {
			prefix = m.styles.Highlight.Render("► ")
			keyStr = m.styles.SelectedKey.Render("[" + opt.Key + "]")
			titleStr = m.styles.Highlight.Render(opt.Title)
		}

		descStr := m.styles.Muted.Render("— " + opt.Description)
		b.WriteString(fmt.Sprintf("%s%s %-14s %s\n", prefix, keyStr, titleStr, descStr))
	}

	b.WriteString("\n" + m.styles.Muted.Render(strings.Repeat("━", max(20, m.width-4))) + "\n")
	b.WriteString(m.styles.Muted.Render("Use [↑/↓] or [1-7] to select destination  •  [Enter] Confirm  •  [Q] Disconnect"))

	return b.String()
}
