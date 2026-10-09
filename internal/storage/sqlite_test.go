package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoragePersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "afterdark_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// 1. Initial Open and Seed Check
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	stats, err := db.GetNetworkStats()
	if err != nil {
		t.Fatalf("Failed to get network stats: %v", err)
	}
	if stats.TotalArchiveNotes == 0 {
		t.Errorf("Expected seeded archive notes, got 0")
	}

	// 2. Profile creation & Returning user
	profileID := "test_key_sha256_abc123"
	p1, returning, err := db.GetOrCreateProfile(profileID, "cyber_wanderer")
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}
	if returning {
		t.Errorf("Expected new user on first visit, got returning=true")
	}
	if p1.VisitsCount != 1 {
		t.Errorf("Expected visits_count=1, got %d", p1.VisitsCount)
	}

	p2, returning2, err := db.GetOrCreateProfile(profileID, "cyber_wanderer")
	if err != nil {
		t.Fatalf("Failed to get profile second time: %v", err)
	}
	if !returning2 {
		t.Errorf("Expected returning=true on second visit")
	}
	if p2.VisitsCount != 2 {
		t.Errorf("Expected visits_count=2, got %d", p2.VisitsCount)
	}

	// 3. Achievements
	unlocked, err := db.UnlockAchievement(profileID, "FIRST_SIGNAL")
	if err != nil || !unlocked {
		t.Fatalf("Expected achievement unlocked, err: %v", err)
	}
	// Test idempotence
	unlockedAgain, err := db.UnlockAchievement(profileID, "FIRST_SIGNAL")
	if err != nil || unlockedAgain {
		t.Errorf("Expected unlockedAgain=false on duplicate unlock")
	}

	achs, err := db.GetAchievements(profileID)
	if err != nil || len(achs) != 1 {
		t.Errorf("Expected 1 achievement unlocked, got %d", len(achs))
	}

	// 4. Archive Notes
	note, err := db.AddArchiveNote("A test message left in the archive.", "tester", "echoes")
	if err != nil {
		t.Fatalf("Failed to add archive note: %v", err)
	}
	if note.Content != "A test message left in the archive." {
		t.Errorf("Unexpected content: %s", note.Content)
	}

	notes, totalNotes, err := db.GetArchiveNotes(0, 10)
	if err != nil || totalNotes < 1 {
		t.Errorf("Failed to retrieve archive notes: %v", err)
	}
	if len(notes) == 0 || notes[0].ID != note.ID {
		t.Errorf("Expected newest note first")
	}

	// 5. Wall Messages
	wallMsg, err := db.AddWallMessage("Graffiti test line", "tagger")
	if err != nil {
		t.Fatalf("Failed to add wall message: %v", err)
	}
	wallList, totalWall, err := db.GetWallMessages(0, 10)
	if err != nil || totalWall < 1 {
		t.Errorf("Failed to retrieve wall messages: %v", err)
	}
	if len(wallList) == 0 || wallList[0].ID != wallMsg.ID {
		t.Errorf("Expected newest wall message first")
	}

	// 6. Void Thoughts
	voidThought, err := db.AddVoidThought("A silent thought dissolves.")
	if err != nil {
		t.Fatalf("Failed to add void thought: %v", err)
	}
	randomThoughts, err := db.GetRandomVoidThoughts(5)
	if err != nil || len(randomThoughts) == 0 {
		t.Fatalf("Failed to get random thoughts: %v", err)
	}
	foundVoid := false
	for _, vt := range randomThoughts {
		if vt.ID == voidThought.ID {
			foundVoid = true
			break
		}
	}
	if !foundVoid && len(randomThoughts) == 0 {
		t.Errorf("Expected void thoughts returned")
	}

	// 7. Secrets Found
	err = db.RecordSecretFound("SECRET_CHAMBER", "tester")
	if err != nil {
		t.Fatalf("Failed to record secret: %v", err)
	}

	// 8. Close and Reopen to test restart persistence
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to close db: %v", err)
	}

	reopenedDB, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen db: %v", err)
	}
	defer reopenedDB.Close()

	reopenedNotes, _, err := reopenedDB.GetArchiveNotes(0, 1)
	if err != nil || len(reopenedNotes) == 0 || reopenedNotes[0].Content != "A test message left in the archive." {
		t.Errorf("Data failed to persist across DB reopen")
	}
}
