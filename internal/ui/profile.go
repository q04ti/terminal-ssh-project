package ui

import (
	"fmt"
	"strings"
	"time"

	"afterdark/internal/security"
	"afterdark/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProfileEditField indicates what is currently being edited.
type ProfileEditField int

const (
	EditNone ProfileEditField = iota
	EditNickname
	EditStatus
)

// ProfileModel manages identity and achievement presentation.
type ProfileModel struct {
	db           *storage.DB
	profile      *storage.Profile
	isPersistent bool // true if derived from SSH public key
	achievements map[string]time.Time
	styles       Styles
	editField    ProfileEditField
	textInput    textinput.Model
	themes       []string
	themeIndex   int
	errMessage   string
	successMsg   string
	width        int
	height       int
}

// NewProfileModel initializes profile view.
func NewProfileModel(db *storage.DB, profile *storage.Profile, isPersistent bool, styles Styles) ProfileModel {
	themes := []string{"cyan", "purple", "green", "amber", "ice"}
	tIdx := 0
	for i, t := range themes {
		if strings.EqualFold(t, profile.AccentTheme) {
			tIdx = i
			break
		}
	}

	ti := textinput.New()
	ti.CharLimit = 50

	m := ProfileModel{
		db:           db,
		profile:      profile,
		isPersistent: isPersistent,
		styles:       styles,
		themes:       themes,
		themeIndex:   tIdx,
		textInput:    ti,
	}
	m.loadAchievements()
	return m
}

func (m *ProfileModel) loadAchievements() {
	if a, err := m.db.GetAchievements(m.profile.ID); err == nil {
		m.achievements = a
	}
}

// SetSize updates dimensions.
func (m *ProfileModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles input events.
func (m *ProfileModel) Update(msg tea.Msg) (tea.Cmd, bool, bool, string) {
	// returns (cmd, returnToHub, themeChanged, newThemeKey)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.errMessage = ""
		m.successMsg = ""

		if m.editField != EditNone {
			switch msg.Type {
			case tea.KeyEnter:
				val := strings.TrimSpace(m.textInput.Value())
				if m.editField == EditNickname {
					clean, err := security.ValidateNickname(val)
					if err != nil {
						m.errMessage = err.Error()
						return nil, false, false, ""
					}
					m.profile.Nickname = clean
					_ = m.db.UpdateProfile(m.profile)
					m.successMsg = "Nickname updated to " + clean
				} else if m.editField == EditStatus {
					clean := security.SanitizeText(val, 60, false)
					m.profile.StatusMessage = clean
					_ = m.db.UpdateProfile(m.profile)
					m.successMsg = "Status updated."
				}
				m.editField = EditNone
				m.textInput.Reset()
				return nil, false, false, ""

			case tea.KeyEsc:
				m.editField = EditNone
				m.textInput.Reset()
				return nil, false, false, ""
			}

			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return cmd, false, false, ""
		}

		// View mode controls
		switch msg.String() {
		case "q", "Q", "esc":
			return nil, true, false, ""

		case "n", "N":
			m.editField = EditNickname
			m.textInput.Placeholder = m.profile.Nickname
			m.textInput.CharLimit = 16
			m.textInput.Focus()
			return nil, false, false, ""

		case "s", "S":
			m.editField = EditStatus
			m.textInput.Placeholder = "enter short status..."
			m.textInput.CharLimit = 50
			m.textInput.Focus()
			return nil, false, false, ""

		case "t", "T":
			m.themeIndex = (m.themeIndex + 1) % len(m.themes)
			newTheme := m.themes[m.themeIndex]
			m.profile.AccentTheme = newTheme
			_ = m.db.UpdateProfile(m.profile)
			m.styles = NewStyles(GetTheme(newTheme))
			m.successMsg = "Theme changed to " + strings.ToUpper(newTheme)
			return nil, false, true, newTheme
		}
	}

	return nil, false, false, ""
}

// View renders the Profile & Achievements UI.
func (m ProfileModel) View() string {
	var b strings.Builder

	header := m.styles.Title.Render("AFTERDARK") + m.styles.Muted.Render(" / ") + m.styles.Prompt.Render("THE PROFILE")
	b.WriteString(header + "\n")
	b.WriteString(m.styles.Muted.Render(strings.Repeat("═", max(20, m.width-4))) + "\n\n")

	if m.successMsg != "" {
		b.WriteString(m.styles.Highlight.Render("✓ "+m.successMsg) + "\n\n")
	}
	if m.errMessage != "" {
		b.WriteString(m.styles.Error.Render("! "+m.errMessage) + "\n\n")
	}

	if m.editField != EditNone {
		promptName := "New Nickname (2-16 chars):"
		if m.editField == EditStatus {
			promptName = "New Status Message:"
		}
		b.WriteString(m.styles.Highlight.Render(promptName) + "\n")
		b.WriteString(m.styles.Prompt.Render("> ") + m.textInput.View() + "\n\n")
		b.WriteString(m.styles.Muted.Render("[Enter] Save  •  [Esc] Cancel"))
		return b.String()
	}

	// Identity Box
	idType := "🔑 Persistent SSH Key ID"
	if !m.isPersistent {
		idType = "⚡ Ephemeral Session ID (Progress resets on reconnect unless connecting via SSH key)"
	}

	idDisplay := m.profile.ID
	if len(idDisplay) > 36 {
		idDisplay = idDisplay[:36] + "..."
	}

	profileBox := []string{
		fmt.Sprintf("Handle:       %s", m.styles.Highlight.Render(m.profile.Nickname)),
		fmt.Sprintf("Status:       %s", m.styles.MenuItem.Render(m.profile.StatusMessage)),
		fmt.Sprintf("Color Theme:  %s", m.styles.Tag.Render(strings.ToUpper(m.profile.AccentTheme))),
		fmt.Sprintf("Identity:     %s", m.styles.Muted.Render(idDisplay)),
		fmt.Sprintf("Mode:         %s", m.styles.Subtitle.Render(idType)),
		fmt.Sprintf("Visits:       %s", m.styles.MenuItem.Render(fmt.Sprintf("%d", m.profile.VisitsCount))),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Theme.Border).
		Padding(0, 2).
		Render(strings.Join(profileBox, "\n"))

	b.WriteString(box + "\n\n")

	// Achievements Section
	b.WriteString(m.styles.Highlight.Render("ACHIEVEMENTS RECORD") + "\n")
	b.WriteString(m.styles.Muted.Render("──────────────────────────────────────────") + "\n")

	unlockedCount := 0
	for _, badge := range storage.AvailableBadges {
		unlockedAt, isUnlocked := m.achievements[badge.Key]
		if isUnlocked {
			unlockedCount++
			ts := unlockedAt.Format("2006-01-02")
			b.WriteString(fmt.Sprintf("%s %s %s - %s\n",
				badge.Icon,
				m.styles.Highlight.Render(badge.Title),
				m.styles.Muted.Render("("+ts+")"),
				m.styles.MenuItem.Render(badge.Desc),
			))
		} else {
			b.WriteString(fmt.Sprintf("🔒 %s - %s\n",
				m.styles.Muted.Render(badge.Title),
				m.styles.Muted.Render(badge.Desc),
			))
		}
	}
	b.WriteString(m.styles.Muted.Render(fmt.Sprintf("Unlocked: %d / %d\n\n", unlockedCount, len(storage.AvailableBadges))))

	b.WriteString(m.styles.Muted.Render(strings.Repeat("═", max(20, m.width-4))) + "\n")
	b.WriteString(m.styles.MenuKey.Render("[N]") + m.styles.Muted.Render(" Nickname  •  ") +
		m.styles.MenuKey.Render("[S]") + m.styles.Muted.Render(" Status  •  ") +
		m.styles.MenuKey.Render("[T]") + m.styles.Muted.Render(" Cycle Theme  •  ") +
		m.styles.MenuKey.Render("[Q/Esc]") + m.styles.Muted.Render(" Return to hub"))

	return b.String()
}
