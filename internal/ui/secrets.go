package ui

import (
	"fmt"
	"strings"

	"afterdark/internal/security"
	"afterdark/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SecretsModel implements the hidden command console and easter egg explorer.
type SecretsModel struct {
	db          *storage.DB
	styles      Styles
	textInput   textinput.Model
	history     []string
	authorNick  string
	unlockedEgg bool
	width       int
	height      int
}

// NewSecretsModel initializes secrets command chamber.
func NewSecretsModel(db *storage.DB, styles Styles, authorNick string) SecretsModel {
	ti := textinput.New()
	ti.Placeholder = "type a command... (try 'help')"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 50

	m := SecretsModel{
		db:         db,
		styles:     styles,
		textInput:  ti,
		authorNick: authorNick,
		history: []string{
			"AFTERDARK SECRETS TERMINAL // RESTRICTED ACCESS",
			"Every node has an echo. Some frequencies still carry whispers.",
			"Type 'help' to inspect command interface.",
			"",
		},
	}
	return m
}

// SetSize updates dimensions.
func (m *SecretsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.textInput.Width = max(20, width-20)
}

// Update processes secret console commands.
func (m *SecretsModel) Update(msg tea.Msg) (tea.Cmd, bool, bool) {
	// returns (cmd, returnToHub, secretUnlocked)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			raw := strings.TrimSpace(m.textInput.Value())
			if raw == "" {
				return nil, false, false
			}

			clean := security.SanitizeText(raw, 64, false)
			m.textInput.Reset()
			secretUnlocked := m.executeCommand(clean)
			return nil, false, secretUnlocked

		case tea.KeyEsc:
			return nil, true, false
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return cmd, false, false
}

func (m *SecretsModel) executeCommand(cmd string) bool {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	m.history = append(m.history, fmt.Sprintf("> %s", cmd))

	unlockedSecret := false

	switch lower {
	case "help":
		m.history = append(m.history,
			"AVAILABLE PROTOCOLS:",
			"  help          - display this transmission index",
			"  whoami        - query active node identity",
			"  beacon        - scan high-frequency emergency transmitters",
			"  matrix        - render stream memory cascade",
			"  constellation - chart the forgotten node coordinates",
			"  glitch        - inspect corrupted core memory sectors",
			"  oracle        - consult the ancient terminal oracle",
			"  chamber       - attempt access to Sub-Basement / Node-0",
			"  clear         - wipe terminal buffer",
			"  exit / hub    - return to central terminal",
			"",
		)

	case "whoami":
		m.history = append(m.history,
			fmt.Sprintf("OPERATOR: %s", m.authorNick),
			"TERMINAL: SSH-CONSOL-VT100",
			"SECURITY CLEARANCE: GUEST // UNVERIFIED ENTITY",
			"",
		)

	case "beacon", "radio", "frequency":
		unlockedSecret = true
		_ = m.db.RecordSecretFound("RADIO_BEACON", m.authorNick)
		m.history = append(m.history,
			"TUNING TO 4625 kHz [UVB-76 BUZZER / THE PIP]...",
			"--------------------------------------------------",
			"SIGNAL LOCK: [████████████████████] 99.4%",
			"TRANSCRIPTION DETECTED:",
			"  'UVB-76 UVB-76 93 882 NAZVANIE 74 14 35 74'",
			"  'To anyone listening through the dark: keep the port open.'",
			"  'The original web never died. It just went underground.'",
			"--------------------------------------------------",
			"",
		)

	case "matrix", "rain":
		unlockedSecret = true
		_ = m.db.RecordSecretFound("MATRIX_CASCADE", m.authorNick)
		m.history = append(m.history,
			"INITIALIZING MEMORY CASCADE BUFFER...",
			"01000001 01000110 01010100 01000101 01010010 01000100 01000001 01010010 01001011",
			"01110011 01101000 01100101 01101100 01101100 01101100 01100101 01110011 01110011",
			"01110011 01110101 01101110 01100100 01100101 01110010 01110111 01101111 01110010",
			"[DECODED]: 'You take the blue cable, you wake up in your feed.'",
			"[DECODED]: 'You stay in AFTERDARK, you see how deep the port goes.'",
			"",
		)

	case "constellation", "stars", "map":
		unlockedSecret = true
		_ = m.db.RecordSecretFound("CONSTELLATION_MAP", m.authorNick)
		m.history = append(m.history,
			"MAPPING HIDDEN UNDERGROUND NODES:",
			"   * (Node-0: 0.0.0.0:2222) ─────── * (Sub-Basement)",
			"           │                             │",
			"           * (The Static Archive)         │",
			"           │                             * (The Echo Chamber)",
			"   * (The Void Horizon) ─────────────────┘",
			"STATUS: 5 persistent clusters active in orbit.",
			"",
		)

	case "glitch", "dump":
		unlockedSecret = true
		_ = m.db.RecordSecretFound("SYSTEM_GLITCH", m.authorNick)
		m.history = append(m.history,
			"CRITICAL: MEMORY PARITY FAULT IN SECTOR 0x7FFE_AFTERDARK",
			"0x0000: 41 46 54 45 52 44 41 52 4B 00 20 22 22 00 00 00",
			"0x0010: FF FF FF FF 00 13 37 DE AD BE EF 00 C0 DE 00 00",
			"[RESTORED PACKET]: 'The real internet was the terminals we visited along the way.'",
			"",
		)

	case "oracle":
		unlockedSecret = true
		_ = m.db.RecordSecretFound("ANCIENT_ORACLE", m.authorNick)
		m.history = append(m.history,
			"ORACLE PONDERS YOUR QUERY...",
			"  'You searched for an escape from modern algorithms,",
			"   and found a room held together by plain text and curiosity.'",
			"  'May your connection never drop and your terminal stay true.'",
			"",
		)

	case "chamber", "subbasement", "gate":
		unlockedSecret = true
		_ = m.db.RecordSecretFound("SECRET_CHAMBER", m.authorNick)
		m.history = append(m.history,
			"ACCESS GRANTED // WELCOME TO SUB-BASEMENT NODE-0",
			"============================================================",
			"You stepped behind the curtain of AFTERDARK.",
			"Here lies the foundation: pure SSH, Go channels, SQLite WAL,",
			"and an unbroken stream of anonymous explorers.",
			"Achievement [EXPLORER] recorded in your profile logs.",
			"============================================================",
			"",
		)

	case "clear", "cls":
		m.history = []string{"TERMINAL BUFFER CLEARED.", ""}

	case "exit", "quit", "hub", "q":
		m.history = append(m.history, "Returning to main terminal hub...")

	default:
		m.history = append(m.history,
			fmt.Sprintf("UNKNOWN PROTOCOL '%s'. Type 'help' for valid commands.", cmd),
			"",
		)
	}

	// Keep history bounded
	if len(m.history) > 100 {
		m.history = m.history[len(m.history)-100:]
	}

	return unlockedSecret
}

// View renders the secrets terminal.
func (m SecretsModel) View() string {
	var b strings.Builder

	header := m.styles.Title.Render("AFTERDARK") + m.styles.Muted.Render(" / ") + m.styles.Prompt.Render("THE SECRETS")
	b.WriteString(header + "\n")
	b.WriteString(m.styles.Muted.Render(strings.Repeat("─", max(20, m.width-4))) + "\n\n")

	// Show last lines of history that fit screen
	maxLines := max(6, m.height-10)
	start := 0
	if len(m.history) > maxLines {
		start = len(m.history) - maxLines
	}

	for _, line := range m.history[start:] {
		if strings.HasPrefix(line, "> ") {
			b.WriteString(m.styles.Highlight.Render(line) + "\n")
		} else if strings.Contains(line, "CRITICAL") || strings.Contains(line, "FAULT") {
			b.WriteString(m.styles.Error.Render(line) + "\n")
		} else if strings.Contains(line, "WELCOME") || strings.Contains(line, "ACCESS GRANTED") {
			b.WriteString(m.styles.Highlight.Render(line) + "\n")
		} else {
			b.WriteString(m.styles.MenuItem.Render(line) + "\n")
		}
	}

	b.WriteString("\n" + m.styles.Muted.Render(strings.Repeat("─", max(20, m.width-4))) + "\n")
	prompt := m.styles.Prompt.Render("COMMAND> ") + m.textInput.View()
	b.WriteString(prompt + "\n")
	b.WriteString(m.styles.Muted.Render("[Enter] Execute command  •  [Esc] Return to hub  •  (try 'beacon', 'matrix', 'chamber')"))

	return b.String()
}
