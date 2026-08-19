package database

import (
	"fmt"
	"database/sql"
)

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

type DataBase struct {
	Name string
	conf Conf
	db *sql.DB
	create bool
}


func DataBaseInit(conf *Conf) (* DataBase, error) {
	switch(conf.Dbtype) {
	case SQLlite:
		return drvSQLliteInit(conf)
	}

	return nil, fmt.Errorf("failed to close database:")
}

func DataBaseExit(ctx *DataBase) error {
	switch(ctx.conf.Dbtype) {
	case SQLlite:
		return drvSQLliteExit(ctx)
	}

	return nil
}


func DataBaseExpoort(ctx *DataBase) (string, error) {
	return "", nil
}

func DataBaseImport(ctx *DataBase) error {
	return nil
}

type Interface interface {
	Init(conf *Conf) (*DataBase, error)
	Exit(ctx *DataBase) error
	CreateDatabase()
	DeleteDataBase()
}
