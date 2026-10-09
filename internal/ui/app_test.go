package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"afterdark/internal/chat"
	"afterdark/internal/config"
	"afterdark/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

func setupTestApp(t *testing.T) (*RootModel, *storage.DB, func()) {
	tmpDir, err := os.MkdirTemp("", "afterdark_ui_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to open db: %v", err)
	}

	cfg := &config.Config{
		ServerName:     "TEST NODE",
		LateNightStart: 0,
		LateNightEnd:   5,
	}

	hub := chat.NewHub(50)
	profile := &storage.Profile{
		ID:            "session:test1234",
		Nickname:      "wanderer_999",
		StatusMessage: "testing ui",
		AccentTheme:   "cyan",
		VisitsCount:   1,
	}

	model := NewRootModel(cfg, db, hub, profile, false, time.Now())

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return model, db, cleanup
}

func TestAppResizeAndBoot(t *testing.T) {
	model, _, cleanup := setupTestApp(t)
	defer cleanup()

	// 1. Send resize message (80x24)
	resModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := resModel.(*RootModel)

	if m.width != 80 || m.height != 24 {
		t.Errorf("Expected dimensions 80x24, got %dx%d", m.width, m.height)
	}

	// 2. View during boot
	view := m.View()
	if !strings.Contains(view, "Welcome to AFTERDARK") {
		t.Errorf("Expected boot screen welcome text in view: %s", view)
	}

	// 3. Complete boot by pressing Enter
	resModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = resModel.(*RootModel)

	if m.state != StateHub {
		t.Errorf("Expected state StateHub after boot Enter, got state=%d", m.state)
	}

	// 4. View during hub
	hubView := m.View()
	if !strings.Contains(hubView, "THE LOBBY") || !strings.Contains(hubView, "THE ARCHIVE") {
		t.Errorf("Expected hub menu options in view: %s", hubView)
	}
}

func TestAppNavigationAndRooms(t *testing.T) {
	model, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Resize and finish boot
	resModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	resModel, _ = resModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := resModel.(*RootModel)

	// Navigate to Lobby (Key "1")
	resModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = resModel.(*RootModel)
	if m.state != StateLobby {
		t.Fatalf("Expected StateLobby, got %d", m.state)
	}
	if !strings.Contains(m.View(), "THE LOBBY") {
		t.Errorf("Expected lobby header in view")
	}

	// Return to Hub with Esc
	resModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = resModel.(*RootModel)
	if m.state != StateHub {
		t.Fatalf("Expected StateHub after Esc, got %d", m.state)
	}

	// Navigate to The Void (Key "4")
	resModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	m = resModel.(*RootModel)
	if m.state != StateVoid {
		t.Fatalf("Expected StateVoid, got %d", m.state)
	}
	if !strings.Contains(m.View(), "THE VOID") {
		t.Errorf("Expected void header in view")
	}

	// Return to Hub
	resModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = resModel.(*RootModel)

	// Navigate to Secrets (Key "7")
	resModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}})
	m = resModel.(*RootModel)
	if m.state != StateSecrets {
		t.Fatalf("Expected StateSecrets, got %d", m.state)
	}

	// Enter secret command "beacon"
	for _, r := range "beacon" {
		resModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = resModel.(*RootModel)
	}
	resModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = resModel.(*RootModel)

	if !strings.Contains(m.View(), "UVB-76") {
		t.Errorf("Expected UVB-76 beacon output in secrets view: %s", m.View())
	}
}
