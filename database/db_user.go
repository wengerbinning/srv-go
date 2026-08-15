package database

import (
	"fmt"
	"database/sql"
)

type UserParam struct {
	username string
	password string
	email    string
}

type UserEntry struct {
	uuid      int64
	username string
	password string
	email    string
	createdstamp string
	updatedstamp string
}

func (e *UserEntry) Uuid() int64      { return e.uuid }
func (e *UserEntry) Username() string  { return e.username }
func (e *UserEntry) Password() string  { return e.password }
func (e *UserEntry) Email() string     { return e.email }

func NewUserParam(username, password, email string) *UserParam {
	return &UserParam{username: username, password: password, email: email}
}

func NewUserEntry(uuid int64, username, password, email string) *UserEntry {
	return &UserEntry{uuid: uuid, username: username, password: password, email: email}
}

func dbUserInitSchema(db *sql.DB) error {
	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		createdstamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		updatedstamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(createTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	return nil
}

func DbUserAdd(db *sql.DB, param *UserParam) (*UserEntry, error) {
	result, err := db.Exec(
		"INSERT INTO users (username, password, email) VALUES (?, ?, ?)",
		param.username, param.password, param.email,);
	if err != nil {
		return nil, fmt.Errorf("failed to add user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}
	return &UserEntry{
		uuid:       id,
		username: param.username,
		password: param.password,
		email: param.email,
	}, nil
}

func DbUserMod(db *sql.DB, entry *UserEntry) (*UserEntry, error) {
	result, err := db.Exec(
		"UPDATE users SET username = ?, password = ?, email = ?, updatedstamp = CURRENT_TIMESTAMP WHERE id = ?",
		entry.username, entry.password, entry.email, entry.uuid,
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
	return &UserEntry{
		uuid:    entry.uuid,
		username: entry.username,
		password: entry.password,
		email:    entry.email,
	}, nil
}

func DbUserDel(db *sql.DB, entry *UserEntry) error {
	result, err := db.Exec("DELETE FROM users WHERE id = ?", entry.uuid)
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

func DbUserSearch(db *sql.DB, param *UserParam) (*UserEntry, error) {
	var argvs []any
	var conds []string
	query := "SELECT id, username, password, email, createdstamp, updatedstamp FROM users WHERE "

	if param.username != "" {
		conds = append(conds, "username = ?")
		argvs = append(argvs, param.username)
	}
	if param.password != "" {
		conds = append(conds, "password = ?")
		argvs = append(argvs, param.password)
	}
	if param.email != "" {
		conds = append(conds, "email = ?")
		argvs = append(argvs, param.email)
	}
	if len(conds) == 0 {
		return nil, fmt.Errorf("at least one search parameter is required")
	}

	for i, c := range conds {
		if i > 0 {
			query += " AND "
		}
		query += c
	}

	row := db.QueryRow(query, argvs...)
	var entry UserEntry
	err := row.Scan(&entry.uuid,
		&entry.username, &entry.email, &entry.password,
		&entry.createdstamp, &entry.updatedstamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &entry, nil
}
