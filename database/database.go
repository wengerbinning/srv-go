package database

import (
	"os"
	"fmt"
	"errors"
	"database/sql"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func EnsureDatabase(dbPath string) (*sql.DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve db path: %w", err)
	}

	_, err = os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Database not found at %s, creating...\n", absPath)
		} else {
			return nil, fmt.Errorf("failed to check database: %w", err)
		}
	} else {
		fmt.Printf("Database already exists at %s\n", absPath)
	}

	db, err := sql.Open("sqlite", absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := dbUserInitSchema(db); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return db, nil
}

func CloseDatabase(db *sql.DB) error {
	if db == nil {
		return nil
	}
	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}


type Type int

const (
    Default Type = iota
    SQLlite
    MySQL
)


type Conf struct {
	Dbtype   Type
	Dbhost string
	Dbport string
	DbPath string
	DbName string
}

type Context struct {
	Name string
	conf Conf
	db   *sql.DB
}

func (ctx *Context) Close() error {
	return CloseDatabase(ctx.db)
}

type Interface interface {
	Init(conf *Conf) (*Context, error)
	Exit(ctx *Context) error
	CreateDatabase()
	DeleteDataBase()
}
