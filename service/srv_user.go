package service

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type User struct {
	ID       int64
	Username string
	Email    string
	Password string
}

type UserStore interface {
	Register(username, email, password string) (*User, error)
	Update(userID int64, username, email, password string) (*User, error)
	Delete(userID int64) error
}

func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return db, nil
}

func initSchema(db *sql.DB) error {
	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(createTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	return nil
}

func RegisterUser(db *sql.DB, username, email, password string) (*User, error) {
	result, err := db.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		username, email, password,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}

	return &User{
		ID:       id,
		Username: username,
		Email:    email,
		Password: password,
	}, nil
}

func UpdateUser(db *sql.DB, userID int64, username, email, password string) (*User, error) {
	result, err := db.Exec(
		"UPDATE users SET username = ?, email = ?, password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		username, email, password, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &User{
		ID:       userID,
		Username: username,
		Email:    email,
		Password: password,
	}, nil
}

func DeleteUser(db *sql.DB, userID int64) error {
	result, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

type sqlUserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) UserStore {
	return &sqlUserStore{db: db}
}

func (s *sqlUserStore) Register(username, email, password string) (*User, error) {
	return RegisterUser(s.db, username, email, password)
}

func (s *sqlUserStore) Update(userID int64, username, email, password string) (*User, error) {
	return UpdateUser(s.db, userID, username, email, password)
}

func (s *sqlUserStore) Delete(userID int64) error {
	return DeleteUser(s.db, userID)
}
