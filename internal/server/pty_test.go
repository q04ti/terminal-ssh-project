package server

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"afterdark/internal/chat"
	"afterdark/internal/config"
	"afterdark/internal/storage"

	gossh "golang.org/x/crypto/ssh"
)

func TestSSHServerInteractivePTYSession(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "afterdark_pty_test_*")
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
		ServerName:     "AFTERDARK TEST NODE",
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

	go func() {
		_ = srv.Start()
	}()

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
		t.Fatalf("Server failed to bind to %s", addr)
	}

	// Connect SSH client
	clientConfig := &gossh.ClientConfig{
		User:            "explorer_alpha",
		Auth:            []gossh.AuthMethod{gossh.Password("none")},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         3 * time.Second,
	}

	client, err := gossh.Dial("tcp", addr, clientConfig)
	if err != nil {
		t.Fatalf("Failed to dial SSH: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Request PTY
	modes := gossh.TerminalModes{
		gossh.ECHO:          0,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 40, 100, modes); err != nil {
		t.Fatalf("Failed to request PTY: %v", err)
	}

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	if err := session.Shell(); err != nil {
		t.Fatalf("Failed to start shell (which launches AFTERDARK directly): %v", err)
	}

	// Wait for boot screen to render
	time.Sleep(400 * time.Millisecond)

	// Press Enter to accept nickname and enter hub
	_, _ = stdinPipe.Write([]byte("\r"))
	time.Sleep(400 * time.Millisecond)

	// Send 'q' to disconnect cleanly
	_, _ = stdinPipe.Write([]byte("q"))
	time.Sleep(400 * time.Millisecond)

	_ = session.Close()
	_ = srv.Close()

	output := stdoutBuf.String()
	if !strings.Contains(output, "AFTERDARK") && !strings.Contains(output, "CONNECTION") {
		t.Errorf("Expected AFTERDARK terminal UI output in SSH stream, got: %s", output)
	}
}
