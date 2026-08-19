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

func (e *UserEntry) Uuid() int64      { return e.uuid     }
func (e *UserEntry) Username() string { return e.username }
func (e *UserEntry) Password() string { return e.password }
func (e *UserEntry) Email() string    { return e.email    }

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

func NewUserParam(username, password, email string) *UserParam {
	return &UserParam{username: username, password: password, email: email}
}

func NewUserEntry(uuid int64, username, password, email string) *UserEntry {
	return &UserEntry{uuid: uuid, username: username, password: password, email: email}
}

func DbUserAdd(ctx *DataBase, param *UserParam) (*UserEntry, error) {
	result, err := ctx.db.Exec(
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

func DbUserMod(ctx *DataBase, entry *UserEntry) (*UserEntry, error) {
	result, err := ctx.db.Exec(
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

func DbUserDel(ctx *DataBase, entry *UserEntry) error {
	result, err := ctx.db.Exec("DELETE FROM users WHERE id = ?", entry.uuid)
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

func DbUserSearch(ctx *DataBase, param *UserParam) ([]*UserEntry, error) {
	var argvs []any
	var conds []string
	var wildcard bool

	addCond := func(field, value string) {
		if value == "*" {
			wildcard = true
			return
		}
		if value != "" {
			conds = append(conds, field+" = ?")
			argvs = append(argvs, value)
		}
	}
	addCond("username", param.username)
	addCond("password", param.password)
	addCond("email", param.email)

	if len(conds) == 0 && !wildcard {
		return nil, fmt.Errorf("at least one search parameter is required")
	}

	query := "SELECT id, username, password, email, createdstamp, updatedstamp FROM users"
	if len(conds) > 0 {
		query += " WHERE "
		for i, c := range conds {
			if i > 0 {
				query += " AND "
			}
			query += c
		}
	}

	rows, err := ctx.db.Query(query, argvs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var entries []*UserEntry
	for rows.Next() {
		var entry UserEntry
		err := rows.Scan(&entry.uuid,
			&entry.username, &entry.password, &entry.email,
			&entry.createdstamp, &entry.updatedstamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		entries = append(entries, &entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}
	return entries, nil
}
