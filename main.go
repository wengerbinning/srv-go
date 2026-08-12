package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wengerbinning/srv/config"
	"github.com/wengerbinning/srv/service"

	_ "modernc.org/sqlite"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	if err := ensureFold(cfg.Storage.SrvPath); err != nil {
		fmt.Fprintf(os.Stderr, "Data dir error: %v\n", err)
		os.Exit(1)
	}

	if err := ensureFold(cfg.Storage.UsrPath); err != nil {
		fmt.Fprintf(os.Stderr, "Data dir error: %v\n", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(cfg.Storage.SrvPath, cfg.Storage.SrvDbName)
	db, err := ensureDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	store := service.NewUserStore(db)

	user, err := store.Register("alice", "alice@example.com", "password123")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Register error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Registered user: id=%d username=%s email=%s\n", user.ID, user.Username, user.Email)

	updated, err := store.Update(user.ID, "alice_new", "alice_new@example.com", "newpassword")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated user: id=%d username=%s email=%s\n", updated.ID, updated.Username, updated.Email)

	if err := store.Delete(user.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Delete error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Deleted user: id=%d\n", user.ID)

	fmt.Println("Service started successfully")
}

func ensureFold(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	return nil
}

func ensureDatabase(dbPath string) (*sql.DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve db path: %w", err)
	}

	if _, err := os.Stat(absPath); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Database not found at %s, creating...\n", absPath)
	} else if err != nil {
		return nil, fmt.Errorf("failed to check database: %w", err)
	} else {
		fmt.Printf("Database already exists at %s\n", absPath)
	}

	db, err := service.Open(absPath)
	if err != nil {
		return nil, err
	}

	fmt.Println("Database initialized successfully")
	return db, nil
}
