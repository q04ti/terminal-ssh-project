package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// DB manages persistent storage backed by SQLite.
type DB struct {
	db *sql.DB
	mu sync.RWMutex
}

// Open initializes the SQLite database, runs migrations, and seeds initial data.
func Open(dbPath string) (*DB, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Configure pool for concurrency
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(10 * time.Minute)

	s := &DB{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	if err := s.seedIfEmpty(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("seeding failed: %w", err)
	}

	return s, nil
}

// Close closes the database connection.
func (s *DB) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

func (s *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS profiles (
			id TEXT PRIMARY KEY,
			nickname TEXT NOT NULL,
			status_message TEXT NOT NULL DEFAULT '',
			accent_theme TEXT NOT NULL DEFAULT 'cyan',
			first_seen DATETIME NOT NULL,
			last_seen DATETIME NOT NULL,
			visits_count INTEGER NOT NULL DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS achievements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id TEXT NOT NULL,
			badge_key TEXT NOT NULL,
			unlocked_at DATETIME NOT NULL,
			UNIQUE(profile_id, badge_key)
		);`,
		`CREATE TABLE IF NOT EXISTS archive_notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			display_id TEXT NOT NULL,
			content TEXT NOT NULL,
			author_alias TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT 'echoes',
			created_at DATETIME NOT NULL,
			is_hidden INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS wall_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			is_hidden INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS void_thoughts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			is_hidden INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS secrets_found (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			secret_key TEXT NOT NULL,
			found_by_alias TEXT NOT NULL,
			found_at DATETIME NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_archive_created ON archive_notes(created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_wall_created ON wall_messages(created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_void_created ON void_thoughts(created_at DESC);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("executing migration query [%s]: %w", q, err)
		}
	}
	return nil
}

func (s *DB) seedIfEmpty() error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM archive_notes").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		seeds := []struct {
			displayID string
			author    string
			category  string
			content   string
		}{
			{
				displayID: "#0001",
				author:    "root_origin",
				category:  "genesis",
				content:   "The wires are humming tonight. If you are reading this through an SSH pipe, know that this corner was built for wanderers who miss when the net was a dark, quiet place.",
			},
			{
				displayID: "#0002",
				author:    "phantom_99",
				category:  "echoes",
				content:   "Somewhere, someone is awake at the exact same second as me, staring into glowing green phosphor.",
			},
			{
				displayID: "#0003",
				author:    "null_diver",
				category:  "transmissions",
				content:   "Remember when websites didn't ask for cookies, phone numbers, or credit cards? Just a blinking cursor and a command prompt.",
			},
			{
				displayID: "#0004",
				author:    "solitary_node",
				category:  "lost-signals",
				content:   "Leave something genuine behind. The web forgets everything fast, but the archive holds on.",
			},
		}

		now := time.Now().Add(-48 * time.Hour)
		for i, seed := range seeds {
			timestamp := now.Add(time.Duration(i*6) * time.Hour)
			_, err := s.db.Exec(
				"INSERT INTO archive_notes (display_id, content, author_alias, category, created_at) VALUES (?, ?, ?, ?, ?)",
				seed.displayID, seed.content, seed.author, seed.category, timestamp,
			)
			if err != nil {
				return err
			}
		}
	}

	var wallCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM wall_messages").Scan(&wallCount)
	if err != nil {
		return err
	}

	if wallCount == 0 {
		wallSeeds := []struct {
			author  string
			content string
		}{
			{"operator", "Welcome to AFTERDARK. Leave your mark on the wall."},
			{"cipher_cat", "Made it through port 2222. The signal is strong."},
			{"midnight_coder", "No javascript. No cookies. Just pure terminal bliss."},
		}
		for i, w := range wallSeeds {
			_, err := s.db.Exec(
				"INSERT INTO wall_messages (content, author, created_at) VALUES (?, ?, ?)",
				w.content, w.author, time.Now().Add(-time.Duration(3-i)*2*time.Hour),
			)
			if err != nil {
				return err
			}
		}
	}

	var voidCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM void_thoughts").Scan(&voidCount)
	if err != nil {
		return err
	}

	if voidCount == 0 {
		voidSeeds := []string{
			"I wonder how many people have stood where I'm standing and felt completely unseen.",
			"The world outside is so loud, but in this terminal window, everything finally goes quiet.",
			"Hope whoever reads this is drinking enough water and sleeping okay tonight.",
		}
		for _, v := range voidSeeds {
			_, err := s.db.Exec(
				"INSERT INTO void_thoughts (content, created_at) VALUES (?, ?)",
				v, time.Now().Add(-12*time.Hour),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetOrCreateProfile retrieves an existing profile or creates a new one.
// Returns profile, whether user was existing (returning), and any error.
func (s *DB) GetOrCreateProfile(id, defaultNick string) (*Profile, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var p Profile
	err := s.db.QueryRow(
		"SELECT id, nickname, status_message, accent_theme, first_seen, last_seen, visits_count FROM profiles WHERE id = ?",
		id,
	).Scan(&p.ID, &p.Nickname, &p.StatusMessage, &p.AccentTheme, &p.FirstSeen, &p.LastSeen, &p.VisitsCount)

	if err == nil {
		// Existing visitor
		p.VisitsCount++
		p.LastSeen = now
		_, _ = s.db.Exec("UPDATE profiles SET visits_count = ?, last_seen = ? WHERE id = ?", p.VisitsCount, now, id)
		return &p, true, nil
	}

	if err != sql.ErrNoRows {
		return nil, false, err
	}

	// New visitor
	p = Profile{
		ID:            id,
		Nickname:      defaultNick,
		StatusMessage: "exploring the forgotten wires",
		AccentTheme:   "cyan",
		FirstSeen:     now,
		LastSeen:      now,
		VisitsCount:   1,
	}

	_, err = s.db.Exec(
		"INSERT INTO profiles (id, nickname, status_message, accent_theme, first_seen, last_seen, visits_count) VALUES (?, ?, ?, ?, ?, ?, ?)",
		p.ID, p.Nickname, p.StatusMessage, p.AccentTheme, p.FirstSeen, p.LastSeen, p.VisitsCount,
	)
	if err != nil {
		return nil, false, err
	}

	return &p, false, nil
}

// UpdateProfile updates nickname, status message, and accent theme.
func (s *DB) UpdateProfile(p *Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		"UPDATE profiles SET nickname = ?, status_message = ?, accent_theme = ?, last_seen = ? WHERE id = ?",
		p.Nickname, p.StatusMessage, p.AccentTheme, time.Now(), p.ID,
	)
	return err
}

// UnlockAchievement unlocks a badge if not already unlocked. Returns true if newly unlocked.
func (s *DB) UnlockAchievement(profileID, badgeKey string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res, err := s.db.Exec(
		"INSERT OR IGNORE INTO achievements (profile_id, badge_key, unlocked_at) VALUES (?, ?, ?)",
		profileID, badgeKey, time.Now(),
	)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// GetAchievements returns a map of unlocked badge keys and their timestamps.
func (s *DB) GetAchievements(profileID string) (map[string]time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT badge_key, unlocked_at FROM achievements WHERE profile_id = ?", profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]time.Time)
	for rows.Next() {
		var key string
		var t time.Time
		if err := rows.Scan(&key, &t); err == nil {
			result[key] = t
		}
	}
	return result, nil
}

// AddArchiveNote inserts a new archive note.
func (s *DB) AddArchiveNote(content, author, category string) (*ArchiveNote, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var nextID int64
	err := s.db.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM archive_notes").Scan(&nextID)
	if err != nil {
		nextID = 1
	}
	displayID := fmt.Sprintf("#%04d", nextID)

	res, err := s.db.Exec(
		"INSERT INTO archive_notes (display_id, content, author_alias, category, created_at) VALUES (?, ?, ?, ?, ?)",
		displayID, content, author, category, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &ArchiveNote{
		ID:          id,
		DisplayID:   displayID,
		Content:     content,
		AuthorAlias: author,
		Category:    category,
		CreatedAt:   now,
	}, nil
}

// GetArchiveNotes fetches paginated notes ordered by newest first.
func (s *DB) GetArchiveNotes(offset, limit int) ([]ArchiveNote, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM archive_notes WHERE is_hidden = 0").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(
		"SELECT id, display_id, content, author_alias, category, created_at FROM archive_notes WHERE is_hidden = 0 ORDER BY id DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, total, err
	}
	defer rows.Close()

	var notes []ArchiveNote
	for rows.Next() {
		var n ArchiveNote
		if err := rows.Scan(&n.ID, &n.DisplayID, &n.Content, &n.AuthorAlias, &n.Category, &n.CreatedAt); err != nil {
			return nil, total, err
		}
		notes = append(notes, n)
	}
	return notes, total, nil
}

// AddWallMessage inserts a new wall message.
func (s *DB) AddWallMessage(content, author string) (*WallMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	res, err := s.db.Exec(
		"INSERT INTO wall_messages (content, author, created_at) VALUES (?, ?, ?)",
		content, author, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &WallMessage{
		ID:        id,
		Content:   content,
		Author:    author,
		CreatedAt: now,
	}, nil
}

// GetWallMessages fetches paginated wall messages ordered by newest first.
func (s *DB) GetWallMessages(offset, limit int) ([]WallMessage, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM wall_messages WHERE is_hidden = 0").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(
		"SELECT id, content, author, created_at FROM wall_messages WHERE is_hidden = 0 ORDER BY id DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, total, err
	}
	defer rows.Close()

	var msgs []WallMessage
	for rows.Next() {
		var m WallMessage
		if err := rows.Scan(&m.ID, &m.Content, &m.Author, &m.CreatedAt); err != nil {
			return nil, total, err
		}
		msgs = append(msgs, m)
	}
	return msgs, total, nil
}

// AddVoidThought inserts a thought into the void. Intentionally no author or IP is stored.
func (s *DB) AddVoidThought(content string) (*VoidThought, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	res, err := s.db.Exec(
		"INSERT INTO void_thoughts (content, created_at) VALUES (?, ?)",
		content, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &VoidThought{
		ID:        id,
		Content:   content,
		CreatedAt: now,
	}, nil
}

// GetRandomVoidThoughts retrieves up to limit randomly chosen void thoughts.
func (s *DB) GetRandomVoidThoughts(limit int) ([]VoidThought, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		"SELECT id, content, created_at FROM void_thoughts WHERE is_hidden = 0 ORDER BY RANDOM() LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var thoughts []VoidThought
	for rows.Next() {
		var t VoidThought
		if err := rows.Scan(&t.ID, &t.Content, &t.CreatedAt); err != nil {
			return nil, err
		}
		thoughts = append(thoughts, t)
	}
	return thoughts, nil
}

// RecordSecretFound records discovery of a secret easter egg.
func (s *DB) RecordSecretFound(secretKey, alias string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		"INSERT INTO secrets_found (secret_key, found_by_alias, found_at) VALUES (?, ?, ?)",
		secretKey, alias, time.Now(),
	)
	return err
}

// GetNetworkStats aggregates server wide metrics.
func (s *DB) GetNetworkStats() (*NetworkStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &NetworkStats{}

	_ = s.db.QueryRow("SELECT COUNT(*) FROM archive_notes WHERE is_hidden = 0").Scan(&stats.TotalArchiveNotes)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM wall_messages WHERE is_hidden = 0").Scan(&stats.TotalWallMessages)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM void_thoughts WHERE is_hidden = 0").Scan(&stats.TotalVoidThoughts)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM secrets_found").Scan(&stats.TotalSecretsFound)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&stats.TotalProfiles)

	rows, err := s.db.Query(
		"SELECT id, secret_key, found_by_alias, found_at FROM secrets_found ORDER BY id DESC LIMIT 5",
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rec SecretRecord
			if err := rows.Scan(&rec.ID, &rec.SecretKey, &rec.FoundByAlias, &rec.FoundAt); err == nil {
				stats.RecentSecrets = append(stats.RecentSecrets, rec)
			}
		}
	}

	return stats, nil
}
