package database

import (
	"os"
	"fmt"
	"errors"
	"database/sql"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DataBaseDriver struct {}

func (d *DataBaseDriver) drvCreateDatabase() {
	fmt.Println("SqliteDriver: CreateDatabase called (auto-created on Init)")
}

func (d *DataBaseDriver) drvDeleteDataBase() {
	fmt.Println("SqliteDriver: DeleteDataBase called")
}

func drvSQLliteInit(conf *Conf) (*DataBase, error) {
	if conf == nil {
		return nil, fmt.Errorf("database config is nil")
	}

	path, err := filepath.Abs(conf.DbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve db path: %w", err)
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	create := false
	file := filepath.Join(path, conf.DbName)
	_, err = os.Stat(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			create = true
		} else {
			return nil, fmt.Errorf("failed to check database: %w", err)
		}
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

	return &DataBase{
		Name: conf.DbName,
		conf: *conf,
		db: db,
		create: create,
	}, nil
}

func drvSQLliteExit(ctx *DataBase) error {
	if ctx == nil || ctx.db == nil {
		return nil
	}

	if err := ctx.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

func SQLliteConf(format string, v ...any) (*Conf, error) {
	path := fmt.Sprintf(format, v...)
	if path == "" {
		return nil, fmt.Errorf("SQLlite database path is empty!")
	}
	file, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve db path: %w", err)
	}
	return &Conf{
		Dbtype: SQLlite,
		DbPath: filepath.Dir(file),
		DbName: filepath.Base(file),
	}, nil
}
