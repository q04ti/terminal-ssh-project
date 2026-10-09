package ui

import (
	"fmt"
	"math/rand"
	"strings"

	"afterdark/internal/security"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BootModel handles onboarding sequence and nickname input.
type BootModel struct {
	textInput     textinput.Model
	generatedNick string
	errMessage    string
	styles        Styles
	width         int
	height        int
	completed     bool
}

// NewBootModel initializes the boot screen.
func NewBootModel(styles Styles) BootModel {
	ti := textinput.New()
	ti.Placeholder = "wanderer"
	ti.Focus()
	ti.CharLimit = 16
	ti.Width = 24

	gen := generateRandomNick()
	ti.Placeholder = gen

	return BootModel{
		textInput:     ti,
		generatedNick: gen,
		styles:        styles,
	}
}

func generateRandomNick() string {
	prefixes := []string{"wanderer", "ghost", "signal", "null", "cipher", "echo", "vector", "static", "glitch", "phosphor"}
	num := rand.Intn(900) + 100
	return fmt.Sprintf("%s_%d", prefixes[rand.Intn(len(prefixes))], num)
}

// Update handles boot screen key events.
func (m *BootModel) Update(msg tea.Msg) (tea.Cmd, bool, string) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			nick := strings.TrimSpace(m.textInput.Value())
			if nick == "" {
				nick = m.generatedNick
			}

			clean, err := security.ValidateNickname(nick)
			if err != nil {
				m.errMessage = err.Error()
				return nil, false, ""
			}

			m.completed = true
			return nil, true, clean

		case tea.KeyCtrlC, tea.KeyEsc:
			// Let caller handle exit if needed
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return cmd, false, ""
}

// View renders the boot experience.
func (m BootModel) View() string {
	var b strings.Builder

	bootLines := []string{
		"CONNECTING TO AFTERDARK...",
		"Establishing encrypted connection...",
		"Synchronizing the network...",
		"Checking for familiar signals...",
		"",
		m.styles.Highlight.Render("CONNECTION ESTABLISHED."),
		"",
		m.styles.Title.Render("Welcome to AFTERDARK."),
		"",
		m.styles.Subtitle.Render("You are connected to a place"),
		m.styles.Subtitle.Render("that technically does not exist."),
		"",
		m.styles.Prompt.Render("What should we call you?"),
		m.styles.Muted.Render(fmt.Sprintf("(Press Enter for %s, or type a custom nickname)", m.generatedNick)),
		"",
		fmt.Sprintf("> %s", m.textInput.View()),
	}

	if m.errMessage != "" {
		bootLines = append(bootLines, "", m.styles.Error.Render("! "+m.errMessage))
	}

	content := strings.Join(bootLines, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Theme.Border).
		Padding(1, 3).
		Render(content)

	b.WriteString(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box))
	return b.String()
}
