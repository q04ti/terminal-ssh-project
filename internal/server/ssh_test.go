package server

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"afterdark/internal/chat"
	"afterdark/internal/config"
	"afterdark/internal/storage"

	gossh "golang.org/x/crypto/ssh"
)

func getFreePort(t *testing.T) int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to bind free port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func TestSSHServerConnectionAndNoShell(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "afterdark_ssh_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	port := getFreePort(t)

	cfg := &config.Config{
		Host:           "127.0.0.1",
		Port:           port,
		DataDir:        tmpDir,
		DBPath:         filepath.Join(tmpDir, "test.db"),
		HostKeyPath:    filepath.Join(tmpDir, "host_key_ed25519"),
		MaxConnections: 10,
		IdleTimeout:    5 * time.Minute,
		ServerName:     "TEST NODE",
	}

	db, err := storage.Open(cfg.DBPath)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	hub := chat.NewHub(10)

	srv, err := New(cfg, db, hub)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- srv.Start()
	}()

	// Wait for server to bind
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ready := false
	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !ready {
		t.Fatalf("Server failed to bind to %s within timeout", addr)
	}

	// Connect as SSH client
	clientConfig := &gossh.ClientConfig{
		User:            "test_explorer",
		Auth:            []gossh.AuthMethod{gossh.Password("none")},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         3 * time.Second,
	}

	client, err := gossh.Dial("tcp", addr, clientConfig)
	if err != nil {
		t.Fatalf("Failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Verify that requesting arbitrary command execution is NOT allowed
	err = session.Run("whoami")
	if err == nil {
		t.Errorf("Arbitrary command execution should have failed/rejected, but returned nil error")
	}

	// Close SSH session and connection
	_ = session.Close()
	_ = client.Close()

	// Stop server cleanly
	if err := srv.Close(); err != nil {
		t.Errorf("Server Close returned error: %v", err)
	}
}
