package server

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"afterdark/internal/chat"
	"afterdark/internal/config"
	"afterdark/internal/storage"
	"afterdark/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	wishbubbletea "github.com/charmbracelet/wish/bubbletea"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
)

// Server encapsulates the SSH daemon.
type Server struct {
	cfg       *config.Config
	db        *storage.DB
	hub       *chat.Hub
	server    *ssh.Server
	startTime time.Time
}

// New creates and configures the SSH server.
func New(cfg *config.Config, db *storage.DB, hub *chat.Hub) (*Server, error) {
	if err := ensureHostKey(cfg.HostKeyPath); err != nil {
		return nil, fmt.Errorf("failed to ensure host key: %w", err)
	}

	s := &Server{
		cfg:       cfg,
		db:        db,
		hub:       hub,
		startTime: time.Now(),
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	teaHandler := func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		// Determine visitor identity
		var identityID string
		isPersistentKey := false

		pubKey := sess.PublicKey()
		if pubKey != nil {
			fp := gossh.FingerprintSHA256(pubKey)
			identityID = "key:" + fp
			isPersistentKey = true
		} else {
			// Ephemeral session
			identityID = "session:" + uuid.New().String()
		}

		user := sess.User()
		if user == "" || strings.ToLower(user) == "root" || strings.ToLower(user) == "admin" {
			user = "wanderer"
		}

		profile, _, err := s.db.GetOrCreateProfile(identityID, user)
		if err != nil {
			log.Error("Failed to get/create profile", "err", err)
			profile = &storage.Profile{
				ID:            identityID,
				Nickname:      user,
				StatusMessage: "exploring",
				AccentTheme:   "cyan",
				VisitsCount:   1,
			}
		}

		model := ui.NewRootModel(s.cfg, s.db, s.hub, profile, isPersistentKey, s.startTime)

		opts := []tea.ProgramOption{
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		}

		return model, opts
	}

	wishServer, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(cfg.HostKeyPath),
		wish.WithIdleTimeout(cfg.IdleTimeout),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			// Accept all public keys to allow key-based persistent identity
			return true
		}),
		wish.WithPasswordAuth(func(ctx ssh.Context, password string) bool {
			// Allow passwordless entry
			return true
		}),
		wish.WithMiddleware(
			wishbubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create wish server: %w", err)
	}

	s.server = wishServer
	return s, nil
}

// Start listens for incoming connections.
func (s *Server) Start() error {
	log.Info("AFTERDARK SSH node starting", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown cleanly stops the SSH server.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info("AFTERDARK SSH node shutting down")
	return s.server.Shutdown(ctx)
}

// Close forcefully stops the SSH server.
func (s *Server) Close() error {
	return s.server.Close()
}

// ensureHostKey generates an ED25519 host key if one does not exist.
func ensureHostKey(keyPath string) error {
	if _, err := os.Stat(keyPath); err == nil {
		return nil // key already exists
	}

	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create host key directory: %w", err)
	}

	_, priv, edErr := ed25519.GenerateKey(rand.Reader)
	if edErr != nil {
		return fmt.Errorf("failed to generate ed25519 key: %w", edErr)
	}

	bytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: bytes,
	}

	file, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open host key file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		return fmt.Errorf("failed to encode pem: %w", err)
	}

	log.Info("Generated persistent ED25519 SSH host key", "path", keyPath)
	return nil
}

// ListenerAddr returns the listening address if started.
func (s *Server) ListenerAddr() net.Addr {
	return nil
}
