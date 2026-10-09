package ui

import (
	"fmt"
	"strings"
	"time"

	"afterdark/internal/chat"
	"afterdark/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NetworkModel displays live system & topology metrics.
type NetworkModel struct {
	db        *storage.DB
	hub       *chat.Hub
	styles    Styles
	stats     *storage.NetworkStats
	roomDist  map[string]int
	online    int
	startTime time.Time
	width     int
	height    int
}

// NewNetworkModel initializes network view.
func NewNetworkModel(db *storage.DB, hub *chat.Hub, styles Styles, startTime time.Time) NetworkModel {
	m := NetworkModel{
		db:        db,
		hub:       hub,
		styles:    styles,
		startTime: startTime,
	}
	m.Refresh()
	return m
}

// Refresh pulls latest data from memory and sqlite.
func (m *NetworkModel) Refresh() {
	if s, err := m.db.GetNetworkStats(); err == nil {
		m.stats = s
	}
	m.online = m.hub.OnlineCount()
	m.roomDist = m.hub.RoomDistribution()
}

// SetSize updates dimensions.
func (m *NetworkModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles keys.
func (m *NetworkModel) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q", "esc":
			return nil, true
		case "r", "R":
			m.Refresh()
			return nil, false
		}
	}
	return nil, false
}

// View renders the network status dashboard.
func (m NetworkModel) View() string {
	var b strings.Builder

	header := m.styles.Title.Render("AFTERDARK") + m.styles.Muted.Render(" / ") + m.styles.Prompt.Render("THE NETWORK")
	b.WriteString(header + "\n")
	b.WriteString(m.styles.Muted.Render(strings.Repeat("═", max(20, m.width-4))) + "\n\n")

	uptime := time.Since(m.startTime).Round(time.Second)

	// ASCII Topology Map
	topoBox := `
   [ NODE-0 ]──( SSH:2222 )──[ ROUTER-CORE ]
        │                            │
   ┌────┴────────────┬───────────────┼───────────────┐
   │                 │               │               │
[ THE LOBBY ]   [ THE WALL ]   [ THE ARCHIVE ]  [ THE VOID ]
  (Live Chat)    (Graffiti)     (Transmissions)  (Silence)
`
	b.WriteString(m.styles.Highlight.Render("NETWORK TOPOLOGY") + "\n")
	b.WriteString(m.styles.Muted.Render(topoBox) + "\n\n")

	// Statistics Grid
	col1 := []string{
		m.styles.Title.Render("TELEMETRY"),
		m.styles.Muted.Render("─────────"),
		fmt.Sprintf("Active Visitors:    %s", m.styles.Highlight.Render(fmt.Sprintf("%d", m.online))),
		fmt.Sprintf("Known Explorers:    %s", m.styles.MenuItem.Render(fmt.Sprintf("%d", m.stats.TotalProfiles))),
		fmt.Sprintf("System Uptime:      %s", m.styles.MenuItem.Render(uptime.String())),
		fmt.Sprintf("Cryptographic Node: %s", m.styles.MenuItem.Render("ED25519-STABLE")),
	}

	col2 := []string{
		m.styles.Title.Render("ARCHIVAL STORE"),
		m.styles.Muted.Render("──────────────"),
		fmt.Sprintf("Archive Transmissions: %s", m.styles.Highlight.Render(fmt.Sprintf("%d", m.stats.TotalArchiveNotes))),
		fmt.Sprintf("Wall Inscriptions:     %s", m.styles.Highlight.Render(fmt.Sprintf("%d", m.stats.TotalWallMessages))),
		fmt.Sprintf("Thoughts in Void:      %s", m.styles.Highlight.Render(fmt.Sprintf("%d", m.stats.TotalVoidThoughts))),
		fmt.Sprintf("Network Secrets Found: %s", m.styles.Highlight.Render(fmt.Sprintf("%d", m.stats.TotalSecretsFound))),
	}

	col1Block := lipgloss.NewStyle().Width(34).Render(strings.Join(col1, "\n"))
	col2Block := lipgloss.NewStyle().Width(36).Render(strings.Join(col2, "\n"))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, col1Block, "   ", col2Block) + "\n\n")

	// Room Distribution
	b.WriteString(m.styles.Highlight.Render("ACTIVE ROOM POPULATION") + "\n")
	roomNames := []string{"Hub", "Lobby", "Archive", "Wall", "Void", "Network", "Profile", "Secrets"}
	var roomItems []string
	for _, r := range roomNames {
		count := m.roomDist[r]
		roomItems = append(roomItems, fmt.Sprintf("%s: %d", r, count))
	}
	b.WriteString(m.styles.Muted.Render(strings.Join(roomItems, "  •  ")) + "\n\n")

	// Recent Discoveries
	if len(m.stats.RecentSecrets) > 0 {
		b.WriteString(m.styles.Highlight.Render("RECENT NETWORK TRANSMISSIONS") + "\n")
		for _, s := range m.stats.RecentSecrets {
			ts := s.FoundAt.Format("15:04")
			b.WriteString(m.styles.Muted.Render(fmt.Sprintf(" [%s] Explorer '%s' tuned into frequency '%s'\n", ts, s.FoundByAlias, s.SecretKey)))
		}
		b.WriteString("\n")
	}

	b.WriteString(m.styles.Muted.Render(strings.Repeat("═", max(20, m.width-4))) + "\n")
	b.WriteString(m.styles.MenuKey.Render("[R]") + m.styles.Muted.Render(" Refresh metrics  •  ") +
		m.styles.MenuKey.Render("[Q/Esc]") + m.styles.Muted.Render(" Return to hub"))

	return b.String()
}
