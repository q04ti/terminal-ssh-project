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

// ArchiveMode indicates whether user is browsing or composing a note.
type ArchiveMode int

const (
	ArchiveBrowse ArchiveMode = iota
	ArchiveWrite
)

// ArchiveModel handles viewing and submitting persistent notes.
type ArchiveModel struct {
	db         *storage.DB
	styles     Styles
	mode       ArchiveMode
	notes      []storage.ArchiveNote
	page       int
	pageSize   int
	totalNotes int64
	textInput  textinput.Model
	category   string
	categories []string
	catIndex   int
	errMessage string
	successMsg string
	authorNick string
	width      int
	height     int
}

// NewArchiveModel initializes the archive view.
func NewArchiveModel(db *storage.DB, styles Styles, authorNick string) ArchiveModel {
	ti := textinput.New()
	ti.Placeholder = "leave an anonymous transmission for future travelers..."
	ti.CharLimit = 300
	ti.Width = 60

	cats := []string{"echoes", "transmissions", "lost-signals", "musings", "glitches"}

	m := ArchiveModel{
		db:         db,
		styles:     styles,
		mode:       ArchiveBrowse,
		page:       0,
		pageSize:   3,
		textInput:  ti,
		categories: cats,
		catIndex:   0,
		category:   cats[0],
		authorNick: authorNick,
	}
	m.loadPage()
	return m
}

func (m *ArchiveModel) loadPage() {
	notes, total, err := m.db.GetArchiveNotes(m.page*m.pageSize, m.pageSize)
	if err == nil {
		m.notes = notes
		m.totalNotes = total
	}
}

// SetSize updates dimensions.
func (m *ArchiveModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.textInput.Width = max(20, width-20)
}

// Update handles keyboard navigation and composition.
func (m *ArchiveModel) Update(msg tea.Msg) (tea.Cmd, bool, bool) {
	// returns (cmd, returnToHub, noteSubmitted)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.errMessage = ""
		m.successMsg = ""

		if m.mode == ArchiveWrite {
			switch msg.Type {
			case tea.KeyEnter:
				raw := strings.TrimSpace(m.textInput.Value())
				if raw == "" {
					m.errMessage = "Note cannot be empty."
					return nil, false, false
				}
				clean := security.SanitizeText(raw, 300, true)
				if clean == "" {
					m.errMessage = "Note contained invalid characters."
					return nil, false, false
				}

				author := m.authorNick
				if author == "" {
					author = "anonymous"
				}

				_, err := m.db.AddArchiveNote(clean, author, m.category)
				if err != nil {
					m.errMessage = "Failed to store note in archive: " + err.Error()
					return nil, false, false
				}

				m.textInput.Reset()
				m.mode = ArchiveBrowse
				m.page = 0
				m.loadPage()
				m.successMsg = "Transmission inscribed into permanent archive."
				return nil, false, true // noteSubmitted triggers achievement

			case tea.KeyEsc:
				m.mode = ArchiveBrowse
				m.textInput.Reset()
				return nil, false, false

			case tea.KeyTab:
				m.catIndex = (m.catIndex + 1) % len(m.categories)
				m.category = m.categories[m.catIndex]
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
			if int64((m.page+1)*m.pageSize) < m.totalNotes {
				m.page++
				m.loadPage()
			}

		case "p", "P", "left", "up":
			if m.page > 0 {
				m.page--
				m.loadPage()
			}

		case "w", "W":
			m.mode = ArchiveWrite
			m.textInput.Focus()
			return nil, false, false
		}
	}

	return nil, false, false
}

// View renders the Archive screen.
func (m ArchiveModel) View() string {
	var b strings.Builder

	// Header
	header := m.styles.Title.Render("AFTERDARK") + m.styles.Muted.Render(" / ") + m.styles.Prompt.Render("THE ARCHIVE")
	b.WriteString(header + "\n")
	b.WriteString(m.styles.Muted.Render(strings.Repeat("─", max(20, m.width-4))) + "\n\n")

	if m.successMsg != "" {
		b.WriteString(m.styles.Highlight.Render("✓ "+m.successMsg) + "\n\n")
	}
	if m.errMessage != "" {
		b.WriteString(m.styles.Error.Render("! "+m.errMessage) + "\n\n")
	}

	if m.mode == ArchiveWrite {
		b.WriteString(m.styles.Highlight.Render("COMPOSE PERMANENT TRANSMISSION") + "\n\n")
		b.WriteString(m.styles.Muted.Render("Category: [Tab to cycle] ") + m.styles.Tag.Render(m.category) + "\n")
		b.WriteString(m.styles.Muted.Render("Author:   ") + m.styles.MenuItem.Render(m.authorNick) + "\n\n")
		b.WriteString(m.styles.Prompt.Render("> ") + m.textInput.View() + "\n\n")
		b.WriteString(m.styles.Muted.Render("[Enter] Inscribe to archive  •  [Tab] Cycle category  •  [Esc] Cancel"))
		return b.String()
	}

	// Browse Mode
	totalPages := (int(m.totalNotes) + m.pageSize - 1) / m.pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	b.WriteString(m.styles.Muted.Render(fmt.Sprintf("Showing %d-%d of %d archived entries (Page %d/%d)",
		m.page*m.pageSize+1,
		min(int(m.totalNotes), (m.page+1)*m.pageSize),
		m.totalNotes,
		m.page+1,
		totalPages,
	)) + "\n\n")

	if len(m.notes) == 0 {
		b.WriteString(m.styles.Muted.Render("No entries found in archive yet. Be the first to leave a transmission.") + "\n")
	} else {
		for _, note := range m.notes {
			idTag := m.styles.Highlight.Render(note.DisplayID)
			dateTag := m.styles.Muted.Render(note.CreatedAt.Format("2006-01-02 15:04"))
			catTag := m.styles.Tag.Render(note.Category)
			authorTag := m.styles.MenuKey.Render("~" + note.AuthorAlias)

			b.WriteString(fmt.Sprintf("%s  %s  %s  %s\n", idTag, dateTag, catTag, authorTag))

			quoteStyle := lipgloss.NewStyle().
				Foreground(m.styles.Theme.Text).
				Italic(true).
				PaddingLeft(2)

			b.WriteString(quoteStyle.Render(fmt.Sprintf("\"%s\"", note.Content)) + "\n\n")
		}
	}

	b.WriteString(m.styles.Muted.Render(strings.Repeat("─", max(20, m.width-4))) + "\n")
	b.WriteString(m.styles.MenuKey.Render("[N]") + m.styles.Muted.Render(" Next page  •  ") +
		m.styles.MenuKey.Render("[P]") + m.styles.Muted.Render(" Previous page  •  ") +
		m.styles.MenuKey.Render("[W]") + m.styles.Muted.Render(" Write note  •  ") +
		m.styles.MenuKey.Render("[Q/Esc]") + m.styles.Muted.Render(" Return to hub"))

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
