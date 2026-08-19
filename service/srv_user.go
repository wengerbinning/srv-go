package service

import (
	"fmt"
	"github.com/wengerbinning/srv/log"
	"github.com/wengerbinning/srv/database"
)

type UserContext struct {
	db  *database.DataBase
	log log.Interface
}

type Scope int

const (
	System Scope = iota
	Service
	Default
)

type User struct {
	Uuid      int64
	Username string
	Email    string
	Password string
	Scope       int
}

type UserInterface interface {
	Registrate(username, email, password string) (*User, error)
	Changerate(uuid int64, username, email, password string) (*User, error)
	Cancellate(uuid int64) error
}

func (s *Context) CreateUserContext() *UserContext {
	return &UserContext{db: s.db, log: s.log}
}

func (s *Context) DeleteUserContext(uctx *UserContext) error {
	uctx.db = nil
	uctx.log = nil
	return nil
}

func dbUser(entry *database.UserEntry) *User {
	return &User{
		Uuid:     entry.Uuid(),
		Username: entry.Username(),
		Email:    entry.Email(),
		Password: entry.Password(),
	}
}

func UserRegistrate(db *database.DataBase, username, email, password string) (*User, error) {
	param := database.NewUserParam(username, password, email)
	entry, err := database.DbUserAdd(db, param)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}
	return dbUser(entry), nil
}

func UserChangerate(db *database.DataBase, uuid int64, username, email, password string) (*User, error) {
	entry := database.NewUserEntry(uuid, username, password, email)
	entry, err := database.DbUserMod(db, entry)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return dbUser(entry), nil
}

func UserCancellate(db *database.DataBase, uuid int64) error {
	entry := database.NewUserEntry(uuid, "", "", "")
	err := database.DbUserDel(db, entry)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *UserContext) Registrate(username, email, password string) (*User, error) {
	return UserRegistrate(s.db, username, email, password)
}

func (s *UserContext) Changerate(uuid int64, username, email, password string) (*User, error) {
	return UserChangerate(s.db, uuid, username, email, password)
}

func (s *UserContext) Cancellate(uuid int64) error {
	return UserCancellate(s.db, uuid)
}
