package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config represents runtime application configuration.
type Config struct {
	Host             string
	Port             int
	DataDir          string
	DBPath           string
	HostKeyPath      string
	MaxConnections   int
	IdleTimeout      time.Duration
	ServerName       string
	LateNightStart   int // hour (0-23)
	LateNightEnd     int // hour (0-23)
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	dataDir := getEnv("DATA_DIR", "data")

	port, err := strconv.Atoi(getEnv("PORT", "2222"))
	if err != nil || port <= 0 || port > 65535 {
		port = 2222
	}

	maxConn, err := strconv.Atoi(getEnv("MAX_CONNECTIONS", "100"))
	if err != nil || maxConn <= 0 {
		maxConn = 100
	}

	idleMinutes, err := strconv.Atoi(getEnv("IDLE_TIMEOUT_MINUTES", "30"))
	if err != nil || idleMinutes <= 0 {
		idleMinutes = 30
	}

	return &Config{
		Host:           getEnv("HOST", "0.0.0.0"),
		Port:           port,
		DataDir:        dataDir,
		DBPath:         getEnv("DB_PATH", filepath.Join(dataDir, "afterdark.db")),
		HostKeyPath:    getEnv("HOST_KEY_PATH", filepath.Join(dataDir, "afterdark_ed25519")),
		MaxConnections: maxConn,
		IdleTimeout:    time.Duration(idleMinutes) * time.Minute,
		ServerName:     getEnv("SERVER_NAME", "AFTERDARK NODE-0"),
		LateNightStart: 0, // 00:00 midnight
		LateNightEnd:   5, // 05:00 AM
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
