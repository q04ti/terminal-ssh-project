package storage

import "time"

// Profile represents a visitor profile, keyed by SSH public key fingerprint or session ID.
type Profile struct {
	ID            string    `json:"id"`
	Nickname      string    `json:"nickname"`
	StatusMessage string    `json:"status_message"`
	AccentTheme   string    `json:"accent_theme"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
	VisitsCount   int       `json:"visits_count"`
}

// Achievement represents an unlocked badge.
type Achievement struct {
	ID         int64     `json:"id"`
	ProfileID  string    `json:"profile_id"`
	BadgeKey   string    `json:"badge_key"`
	Title      string    `json:"title"`
	Desc       string    `json:"desc"`
	UnlockedAt time.Time `json:"unlocked_at"`
}

// BadgeDefinition provides metadata for achievements.
type BadgeDefinition struct {
	Key   string
	Title string
	Desc  string
	Icon  string
}

var AvailableBadges = []BadgeDefinition{
	{
		Key:   "FIRST_SIGNAL",
		Title: "FIRST SIGNAL",
		Desc:  "Transmitted your first broadcast in the lobby.",
		Icon:  "📡",
	},
	{
		Key:   "ARCHIVIST",
		Title: "ARCHIVIST",
		Desc:  "Preserved a transmission in the permanent archive.",
		Icon:  "📜",
	},
	{
		Key:   "WALL_WRITER",
		Title: "GRAFFITI TAG",
		Desc:  "Carved your mark into the public wall.",
		Icon:  "✒️",
	},
	{
		Key:   "VOID_DRIFTER",
		Title: "VOID DRIFTER",
		Desc:  "Cast an anonymous thought into the silent void.",
		Icon:  "🌌",
	},
	{
		Key:   "EXPLORER",
		Title: "EXPLORER",
		Desc:  "Discovered a hidden chamber or system secret.",
		Icon:  "🗝️",
	},
	{
		Key:   "NIGHT_OWL",
		Title: "NIGHT OWL",
		Desc:  "Connected to AFTERDARK during midnight hours (00:00-05:00).",
		Icon:  "🦉",
	},
	{
		Key:   "OLD_SOUL",
		Title: "OLD SOUL",
		Desc:  "Returned to AFTERDARK across multiple sessions using a recognized SSH key.",
		Icon:  "⏳",
	},
}

// ArchiveNote represents a persistent note in the archive.
type ArchiveNote struct {
	ID          int64     `json:"id"`
	DisplayID   string    `json:"display_id"`
	Content     string    `json:"content"`
	AuthorAlias string    `json:"author_alias"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// WallMessage represents a public wall graffiti post.
type WallMessage struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// VoidThought represents a purely anonymous thought left in the void.
type VoidThought struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// SecretRecord tracks discoveries of secrets across the network.
type SecretRecord struct {
	ID           int64     `json:"id"`
	SecretKey    string    `json:"secret_key"`
	FoundByAlias string    `json:"found_by_alias"`
	FoundAt      time.Time `json:"found_at"`
}

// NetworkStats holds aggregated metrics for the network dashboard.
type NetworkStats struct {
	TotalArchiveNotes int64
	TotalWallMessages int64
	TotalVoidThoughts int64
	TotalSecretsFound int64
	TotalProfiles     int64
	RecentSecrets     []SecretRecord
}
