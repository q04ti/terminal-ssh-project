package ui

import (
	"fmt"
	"time"

	"afterdark/internal/chat"
	"afterdark/internal/config"
	"afterdark/internal/security"
	"afterdark/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AppState enumerates room navigation states.
type AppState int

const (
	StateBoot AppState = iota
	StateHub
	StateLobby
	StateArchive
	StateWall
	StateVoid
	StateNetwork
	StateProfile
	StateSecrets
	StateDisconnecting
)

// ToastMessage represents an achievement or alert banner.
type ToastMessage struct {
	Text      string
	ExpiresAt time.Time
}

// RootModel is the root Bubble Tea model for an SSH session.
type RootModel struct {
	cfg          *config.Config
	db           *storage.DB
	hub          *chat.Hub
	client       *chat.Client
	profile      *storage.Profile
	isPersistent bool
	state        AppState
	styles       Styles
	rateLimiter  *security.RateLimiter
	startTime    time.Time
	toast        *ToastMessage

	width  int
	height int

	// Views
	boot        BootModel
	hubView     HubViewModel
	lobby       LobbyModel
	archive     ArchiveModel
	wall        WallModel
	void        VoidModel
	network     NetworkModel
	profileView ProfileModel
	secrets     SecretsModel
}

// ChatIncomingMsg wraps a broadcast message received from chat hub.
type ChatIncomingMsg struct {
	Message chat.Message
}

// ToastExpireMsg triggers toast removal.
type ToastExpireMsg struct{}

// NewRootModel creates the session root model.
func NewRootModel(
	cfg *config.Config,
	db *storage.DB,
	hub *chat.Hub,
	profile *storage.Profile,
	isPersistent bool,
	startTime time.Time,
) *RootModel {
	theme := GetTheme(profile.AccentTheme)
	styles := NewStyles(theme)

	// Register with chat hub
	client := hub.Register(profile.ID, profile.Nickname, profile.StatusMessage, "Boot")

	rl := security.NewRateLimiter(5, 2*time.Second) // 5 burst, replenish every 2s

	boot := NewBootModel(styles)
	hubView := NewHubViewModel(styles, cfg.ServerName)
	lobby := NewLobbyModel(styles, hub.GetRecentHistory())
	archive := NewArchiveModel(db, styles, profile.Nickname)
	wall := NewWallModel(db, styles, profile.Nickname)
	voidM := NewVoidModel(db, styles)
	network := NewNetworkModel(db, hub, styles, startTime)
	profView := NewProfileModel(db, profile, isPersistent, styles)
	secrets := NewSecretsModel(db, styles, profile.Nickname)

	m := &RootModel{
		cfg:          cfg,
		db:           db,
		hub:          hub,
		client:       client,
		profile:      profile,
		isPersistent: isPersistent,
		state:        StateBoot,
		styles:       styles,
		rateLimiter:  rl,
		startTime:    startTime,

		boot:        boot,
		hubView:     hubView,
		lobby:       lobby,
		archive:     archive,
		wall:        wall,
		void:        voidM,
		network:     network,
		profileView: profView,
		secrets:     secrets,
	}

	// Check time-based and return visitor achievements
	m.checkInitialAchievements()

	return m
}

func (m *RootModel) checkInitialAchievements() {
	now := time.Now()
	hour := now.Hour()
	if hour >= m.cfg.LateNightStart && hour < m.cfg.LateNightEnd {
		m.tryUnlockAchievement("NIGHT_OWL", "NIGHT OWL 🦉 (Midnight hours)")
	}

	if m.isPersistent && m.profile.VisitsCount > 1 {
		m.tryUnlockAchievement("OLD_SOUL", "OLD SOUL ⏳ (Recognized returning visitor)")
	}
}

func (m *RootModel) tryUnlockAchievement(badgeKey, title string) {
	unlocked, err := m.db.UnlockAchievement(m.profile.ID, badgeKey)
	if err == nil && unlocked {
		m.triggerToast(fmt.Sprintf("✦ ACHIEVEMENT UNLOCKED: %s", title))
	}
}

func (m *RootModel) triggerToast(text string) {
	m.toast = &ToastMessage{
		Text:      text,
		ExpiresAt: time.Now().Add(5 * time.Second),
	}
}

func waitForChatMessage(ch chan chat.Message) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return ChatIncomingMsg{Message: msg}
	}
}

// Init initializes tea event loop.
func (m *RootModel) Init() tea.Cmd {
	return tea.Batch(
		waitForChatMessage(m.client.SendCh),
	)
}

// Update handles state transitions, window sizing, and input.
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.boot.width = msg.Width
		m.boot.height = msg.Height
		m.hubView.SetSize(msg.Width, msg.Height)
		m.lobby.SetSize(msg.Width, msg.Height)
		m.archive.SetSize(msg.Width, msg.Height)
		m.wall.SetSize(msg.Width, msg.Height)
		m.void.SetSize(msg.Width, msg.Height)
		m.network.SetSize(msg.Width, msg.Height)
		m.profileView.SetSize(msg.Width, msg.Height)
		m.secrets.SetSize(msg.Width, msg.Height)
		return m, nil

	case ChatIncomingMsg:
		m.lobby.AddMessage(msg.Message)
		// Continue listening for next chat message
		cmds = append(cmds, waitForChatMessage(m.client.SendCh))
		return m, tea.Batch(cmds...)

	case ToastExpireMsg:
		m.toast = nil
		return m, nil
	}

	// Update active online count & users list
	m.hubView.SetOnlineUsers(m.hub.OnlineCount())
	m.lobby.SetUsers(m.hub.GetOnlineUsers())

	// Delegate to current state view
	switch m.state {
	case StateBoot:
		cmd, done, nickname := m.boot.Update(msg)
		cmds = append(cmds, cmd)
		if done {
			m.profile.Nickname = nickname
			_ = m.db.UpdateProfile(m.profile)
			m.hub.UpdatePresence(m.profile.ID, nickname, m.profile.StatusMessage)
			m.archive.authorNick = nickname
			m.wall.authorNick = nickname
			m.secrets.authorNick = nickname
			m.state = StateHub
			m.hub.UpdateRoom(m.profile.ID, "Hub")
		}

	case StateHub:
		roomKey, selected := m.hubView.Update(msg)
		if selected {
			switch roomKey {
			case "Lobby":
				m.state = StateLobby
				m.hub.UpdateRoom(m.profile.ID, "Lobby")
			case "Archive":
				m.state = StateArchive
				m.hub.UpdateRoom(m.profile.ID, "Archive")
				m.archive.loadPage()
			case "Wall":
				m.state = StateWall
				m.hub.UpdateRoom(m.profile.ID, "Wall")
				m.wall.loadPage()
			case "Void":
				m.state = StateVoid
				m.hub.UpdateRoom(m.profile.ID, "Void")
				m.void.revealWhispers()
			case "Network":
				m.state = StateNetwork
				m.hub.UpdateRoom(m.profile.ID, "Network")
				m.network.Refresh()
			case "Profile":
				m.state = StateProfile
				m.hub.UpdateRoom(m.profile.ID, "Profile")
				m.profileView.loadAchievements()
			case "Secrets":
				m.state = StateSecrets
				m.hub.UpdateRoom(m.profile.ID, "Secrets")
			case "Disconnect":
				m.hub.Unregister(m.profile.ID)
				m.state = StateDisconnecting
				return m, tea.Quit
			}
		}

	case StateLobby:
		cmd, chatMsg, returnToHub := m.lobby.Update(msg)
		cmds = append(cmds, cmd)
		if returnToHub {
			m.state = StateHub
			m.hub.UpdateRoom(m.profile.ID, "Hub")
		} else if chatMsg != nil {
			// Rate limit check
			if !m.rateLimiter.Allow(m.profile.ID) {
				m.lobby.errMessage = "Transmit limit reached. Please wait a moment."
			} else {
				m.hub.BroadcastUser(m.profile.Nickname, *chatMsg, string(m.styles.Theme.Primary))
				m.tryUnlockAchievement("FIRST_SIGNAL", "FIRST SIGNAL 📡")
			}
		}

	case StateArchive:
		cmd, returnToHub, noteSubmitted := m.archive.Update(msg)
		cmds = append(cmds, cmd)
		if returnToHub {
			m.state = StateHub
			m.hub.UpdateRoom(m.profile.ID, "Hub")
		}
		if noteSubmitted {
			m.tryUnlockAchievement("ARCHIVIST", "ARCHIVIST 📜")
		}

	case StateWall:
		cmd, returnToHub, msgSubmitted := m.wall.Update(msg)
		cmds = append(cmds, cmd)
		if returnToHub {
			m.state = StateHub
			m.hub.UpdateRoom(m.profile.ID, "Hub")
		}
		if msgSubmitted {
			m.tryUnlockAchievement("WALL_WRITER", "GRAFFITI TAG ✒️")
		}

	case StateVoid:
		cmd, returnToHub, thoughtCast := m.void.Update(msg)
		cmds = append(cmds, cmd)
		if returnToHub {
			m.state = StateHub
			m.hub.UpdateRoom(m.profile.ID, "Hub")
		}
		if thoughtCast {
			m.tryUnlockAchievement("VOID_DRIFTER", "VOID DRIFTER 🌌")
		}

	case StateNetwork:
		cmd, returnToHub := m.network.Update(msg)
		cmds = append(cmds, cmd)
		if returnToHub {
			m.state = StateHub
			m.hub.UpdateRoom(m.profile.ID, "Hub")
		}

	case StateProfile:
		cmd, returnToHub, themeChanged, newTheme := m.profileView.Update(msg)
		cmds = append(cmds, cmd)
		if returnToHub {
			m.state = StateHub
			m.hub.UpdateRoom(m.profile.ID, "Hub")
		}
		if themeChanged {
			m.styles = NewStyles(GetTheme(newTheme))
			m.hubView.styles = m.styles
			m.lobby.styles = m.styles
			m.archive.styles = m.styles
			m.wall.styles = m.styles
			m.void.styles = m.styles
			m.network.styles = m.styles
			m.profileView.styles = m.styles
			m.secrets.styles = m.styles
		}

	case StateSecrets:
		cmd, returnToHub, secretUnlocked := m.secrets.Update(msg)
		cmds = append(cmds, cmd)
		if returnToHub {
			m.state = StateHub
			m.hub.UpdateRoom(m.profile.ID, "Hub")
		}
		if secretUnlocked {
			m.tryUnlockAchievement("EXPLORER", "EXPLORER 🗝️")
		}
	}

	return m, tea.Batch(cmds...)
}

// View delegates rendering to active screen.
func (m *RootModel) View() string {
	if m.state == StateDisconnecting {
		return m.styles.Muted.Render("\nConnection severed. The line goes dead.\nGoodbye, explorer.\n\n")
	}

	var content string
	switch m.state {
	case StateBoot:
		content = m.boot.View()
	case StateHub:
		content = m.hubView.View()
	case StateLobby:
		content = m.lobby.View()
	case StateArchive:
		content = m.archive.View()
	case StateWall:
		content = m.wall.View()
	case StateVoid:
		content = m.void.View()
	case StateNetwork:
		content = m.network.View()
	case StateProfile:
		content = m.profileView.View()
	case StateSecrets:
		content = m.secrets.View()
	}

	if m.toast != nil && time.Now().Before(m.toast.ExpiresAt) {
		toastBar := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.styles.Theme.Highlight).
			Background(m.styles.Theme.BgAlt).
			Border(lipgloss.NormalBorder()).
			BorderForeground(m.styles.Theme.Highlight).
			Padding(0, 2).
			Render(m.toast.Text)

		content = toastBar + "\n\n" + content
	}

	return content
}
