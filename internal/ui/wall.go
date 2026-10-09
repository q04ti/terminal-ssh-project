package ui

import (
	"fmt"
	"strings"

	"afterdark/internal/security"
	"afterdark/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// WallMode indicates browsing or graffiti carving mode.
type WallMode int

const (
	WallBrowse WallMode = iota
	WallWrite
)

// WallModel handles viewing and posting to the public graffiti wall.
type WallModel struct {
	db         *storage.DB
	styles     Styles
	mode       WallMode
	messages   []storage.WallMessage
	page       int
	pageSize   int
	totalMsgs  int64
	textInput  textinput.Model
	errMessage string
	successMsg string
	authorNick string
	width      int
	height     int
}

// NewWallModel initializes the wall view.
func NewWallModel(db *storage.DB, styles Styles, authorNick string) WallModel {
	ti := textinput.New()
	ti.Placeholder = "carve your tag or short message on the wall..."
	ti.CharLimit = 200
	ti.Width = 60

	m := WallModel{
		db:         db,
		styles:     styles,
		mode:       WallBrowse,
		page:       0,
		pageSize:   5,
		textInput:  ti,
		authorNick: authorNick,
	}
	m.loadPage()
	return m
}

func (m *WallModel) loadPage() {
	msgs, total, err := m.db.GetWallMessages(m.page*m.pageSize, m.pageSize)
	if err == nil {
		m.messages = msgs
		m.totalMsgs = total
	}
}

// SetSize updates dimensions.
func (m *WallModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.textInput.Width = max(20, width-20)
}

// Update handles keyboard navigation and carving messages.
func (m *WallModel) Update(msg tea.Msg) (tea.Cmd, bool, bool) {
	// returns (cmd, returnToHub, msgSubmitted)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.errMessage = ""
		m.successMsg = ""

		if m.mode == WallWrite {
			switch msg.Type {
			case tea.KeyEnter:
				raw := strings.TrimSpace(m.textInput.Value())
				if raw == "" {
					m.errMessage = "Tag cannot be empty."
					return nil, false, false
				}
				clean := security.SanitizeText(raw, 200, false)
				if clean == "" {
					m.errMessage = "Message contained invalid characters."
					return nil, false, false
				}

				author := m.authorNick
				if author == "" {
					author = "wanderer"
				}

				_, err := m.db.AddWallMessage(clean, author)
				if err != nil {
					m.errMessage = "Failed to carve message: " + err.Error()
					return nil, false, false
				}

				m.textInput.Reset()
				m.mode = WallBrowse
				m.page = 0
				m.loadPage()
				m.successMsg = "Tag permanently carved into the wall."
				return nil, false, true // msgSubmitted triggers achievement

			case tea.KeyEsc:
				m.mode = WallBrowse
				m.textInput.Reset()
				return nil, false, false
			}

			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return cmd, false, false
		}

		// Browse mode controls
		switch msg.String() {
		case "q", "Q", "esc":
			return nil, true, false // Return to hub

		case "n", "N", "right", "down":
			if int64((m.page+1)*m.pageSize) < m.totalMsgs {
				m.page++
				m.loadPage()
			}

		case "p", "P", "left", "up":
			if m.page > 0 {
				m.page--
				m.loadPage()
			}

		case "w", "W":
			m.mode = WallWrite
			m.textInput.Focus()
			return nil, false, false
		}
	}

	return nil, false, false
}

// View renders the Wall screen.
func (m WallModel) View() string {
	var b strings.Builder

	header := m.styles.Title.Render("AFTERDARK") + m.styles.Muted.Render(" / ") + m.styles.Prompt.Render("THE WALL")
	b.WriteString(header + "\n")
	b.WriteString(m.styles.Muted.Render(strings.Repeat("═", max(20, m.width-4))) + "\n\n")

	if m.successMsg != "" {
		b.WriteString(m.styles.Highlight.Render("✓ "+m.successMsg) + "\n\n")
	}
	if m.errMessage != "" {
		b.WriteString(m.styles.Error.Render("! "+m.errMessage) + "\n\n")
	}

	if m.mode == WallWrite {
		b.WriteString(m.styles.Highlight.Render("CARVE YOUR MARK ON THE WALL") + "\n\n")
		b.WriteString(m.styles.Muted.Render("Author: ") + m.styles.MenuItem.Render(m.authorNick) + "\n\n")
		b.WriteString(m.styles.Prompt.Render("Tag: ") + m.textInput.View() + "\n\n")
		b.WriteString(m.styles.Muted.Render("[Enter] Carve permanently  •  [Esc] Cancel"))
		return b.String()
	}

	totalPages := (int(m.totalMsgs) + m.pageSize - 1) / m.pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	b.WriteString(m.styles.Muted.Render(fmt.Sprintf("Public Graffiti Wall • %d inscriptions recorded (Page %d/%d)",
		m.totalMsgs,
		m.page+1,
		totalPages,
	)) + "\n\n")

	if len(m.messages) == 0 {
		b.WriteString(m.styles.Muted.Render("The wall is bare concrete. Be the first to carve your mark.") + "\n")
	} else {
		for _, msg := range m.messages {
			dateTag := m.styles.Muted.Render(msg.CreatedAt.Format("01/02 15:04"))
			authorTag := m.styles.Highlight.Render("<" + msg.Author + ">")
			text := m.styles.MenuItem.Render(msg.Content)

			b.WriteString(fmt.Sprintf("%s %s %s\n", dateTag, authorTag, text))
		}
	}

	b.WriteString("\n" + m.styles.Muted.Render(strings.Repeat("═", max(20, m.width-4))) + "\n")
	b.WriteString(m.styles.MenuKey.Render("[N]") + m.styles.Muted.Render(" Next page  •  ") +
		m.styles.MenuKey.Render("[P]") + m.styles.Muted.Render(" Previous page  •  ") +
		m.styles.MenuKey.Render("[W]") + m.styles.Muted.Render(" Carve tag  •  ") +
		m.styles.MenuKey.Render("[Q/Esc]") + m.styles.Muted.Render(" Return to hub"))

	return b.String()
}
