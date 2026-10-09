package ui

import (
	"fmt"
	"strings"

	"afterdark/internal/security"
	"afterdark/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// VoidModel handles the quiet anonymous thought room.
type VoidModel struct {
	db         *storage.DB
	styles     Styles
	textInput  textinput.Model
	whispers   []storage.VoidThought
	errMessage string
	successMsg string
	width      int
	height     int
}

// NewVoidModel initializes the void view.
func NewVoidModel(db *storage.DB, styles Styles) VoidModel {
	ti := textinput.New()
	ti.Placeholder = "leave one thought behind..."
	ti.Focus()
	ti.CharLimit = 250
	ti.Width = 55

	m := VoidModel{
		db:        db,
		styles:    styles,
		textInput: ti,
	}
	m.revealWhispers()
	return m
}

func (m *VoidModel) revealWhispers() {
	thoughts, err := m.db.GetRandomVoidThoughts(3)
	if err == nil {
		m.whispers = thoughts
	}
}

// SetSize updates dimensions.
func (m *VoidModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.textInput.Width = max(20, width-20)
}

// Update handles interactions in the void.
func (m *VoidModel) Update(msg tea.Msg) (tea.Cmd, bool, bool) {
	// returns (cmd, returnToHub, thoughtCast)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.errMessage = ""
		m.successMsg = ""

		switch msg.Type {
		case tea.KeyEnter:
			val := strings.TrimSpace(m.textInput.Value())
			if val == "" {
				return nil, false, false
			}
			clean := security.SanitizeText(val, 250, false)
			if clean == "" {
				m.errMessage = "Invalid characters."
				return nil, false, false
			}

			_, err := m.db.AddVoidThought(clean)
			if err != nil {
				m.errMessage = "The void rejected your thought: " + err.Error()
				return nil, false, false
			}

			m.textInput.Reset()
			m.successMsg = "Your thought dissolved into the darkness. No records were kept."
			m.revealWhispers()
			return nil, false, true // thoughtCast triggers VOID_DRIFTER achievement

		case tea.KeyEsc:
			return nil, true, false
		}

		switch msg.String() {
		case "q", "Q":
			if m.textInput.Value() == "" {
				return nil, true, false
			}
		case "r", "R":
			if m.textInput.Value() == "" {
				m.revealWhispers()
				m.successMsg = "Whispers from the void surfaced."
				return nil, false, false
			}
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return cmd, false, false
}

// View renders the Void interface.
func (m VoidModel) View() string {
	var b strings.Builder

	header := m.styles.Title.Render("AFTERDARK") + m.styles.Muted.Render(" / ") + m.styles.Prompt.Render("THE VOID")
	b.WriteString(header + "\n\n")

	introLines := []string{
		m.styles.Highlight.Render("THE VOID"),
		"",
		m.styles.MenuItem.Render("Leave one thought behind."),
		m.styles.Subtitle.Render("Nobody needs to know who you are."),
		m.styles.Muted.Render("(No identity, no IP address, no SSH key is ever stored here.)"),
		"",
	}
	b.WriteString(strings.Join(introLines, "\n") + "\n")

	if m.successMsg != "" {
		b.WriteString(m.styles.Highlight.Render("· "+m.successMsg) + "\n\n")
	}
	if m.errMessage != "" {
		b.WriteString(m.styles.Error.Render("! "+m.errMessage) + "\n\n")
	}

	b.WriteString(m.styles.Prompt.Render("> ") + m.textInput.View() + "\n\n")

	b.WriteString(m.styles.Muted.Render("── Whispers from the dark ────────────────") + "\n\n")

	if len(m.whispers) == 0 {
		b.WriteString(m.styles.Muted.Render("   The silence here is absolute.") + "\n\n")
	} else {
		for _, w := range m.whispers {
			quoteStyle := lipgloss.NewStyle().
				Foreground(m.styles.Theme.Text).
				Italic(true).
				PaddingLeft(3)
			b.WriteString(quoteStyle.Render(fmt.Sprintf("• \"%s\"", w.Content)) + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(m.styles.Muted.Render("──────────────────────────────────────────") + "\n")
	b.WriteString(m.styles.MenuKey.Render("[Enter]") + m.styles.Muted.Render(" Release thought  •  ") +
		m.styles.MenuKey.Render("[R]") + m.styles.Muted.Render(" Reveal other whispers  •  ") +
		m.styles.MenuKey.Render("[Esc/Q]") + m.styles.Muted.Render(" Return to hub"))

	return b.String()
}
