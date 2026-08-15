package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type SqliteDriver struct{}

func NewSqliteDriver() *SqliteDriver {
	return &SqliteDriver{}
}

func SQLliteConf(format string, v ...any) (*Conf, error) {
	path := fmt.Sprintf(format, v...)
	if path == "" {
		return nil, fmt.Errorf("database path is empty")
	}
	file, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve db path: %w", err)
	}
	return &Conf{ Dbtype: SQLlite,
		DbPath: filepath.Dir(file), DbName: filepath.Base(file), }, nil
}


func (d *SqliteDriver) Init(conf *Conf) (*Context, error) {
	if conf == nil {
		return nil, fmt.Errorf("database config is nil")
	}

	path, err := filepath.Abs(conf.DbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve db path: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	file := filepath.Join(path, conf.DbName)
	_, err = os.Stat(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Database not found at %s, creating...\n", file)
		} else {
			return nil, fmt.Errorf("failed to check database: %w", err)
		}
	} else {
		fmt.Printf("Database already exists at %s\n", file)
	}

	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := dbUserInitSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return &Context{
		Name: conf.DbName,
		conf: *conf,
		db:   db,
	}, nil
}

func (d *SqliteDriver) Exit(ctx *Context) error {
	if ctx == nil || ctx.db == nil {
		return nil
	}
	if err := ctx.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

func (d *SqliteDriver) CreateDatabase() {
	fmt.Println("SqliteDriver: CreateDatabase called (auto-created on Init)")
}

func (d *SqliteDriver) DeleteDataBase() {
	fmt.Println("SqliteDriver: DeleteDataBase called")
}
