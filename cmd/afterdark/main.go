package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"afterdark/internal/chat"
	"afterdark/internal/config"
	"afterdark/internal/server"
	"afterdark/internal/storage"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func main() {
	log.SetLevel(log.InfoLevel)
	log.Info("AFTERDARK: initializing network node...")

	cfg := config.Load()

	// 1. Storage
	db, err := storage.Open(cfg.DBPath)
	if err != nil {
		log.Fatal("Failed to open SQLite database", "err", err)
	}
	defer db.Close()
	log.Info("Database initialized and verified", "path", cfg.DBPath)

	// 2. Chat & Presence Hub
	hub := chat.NewHub(100)
	log.Info("Multiplayer chat hub online")

	// 3. SSH Server
	srv, err := server.New(cfg, db, hub)
	if err != nil {
		log.Fatal("Failed to initialize SSH server", "err", err)
	}

	// 4. Start Server
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, ssh.ErrServerClosed) && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("SSH server failed unexpectedly", "err", err)
		}
	}()

	log.Info("=====================================================")
	log.Info("AFTERDARK NODE IS ONLINE AND AWAITING EXPLORERS")
	log.Info("Connect via: ssh 127.0.0.1 -p " + os.Getenv("PORT"))
	log.Info("=====================================================")

	<-done
	log.Info("Interrupt signal received. Beginning graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Error during SSH server shutdown", "err", err)
	}

	log.Info("AFTERDARK node terminated cleanly. Disconnected.")
}
