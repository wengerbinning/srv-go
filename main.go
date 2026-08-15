package main

import (
	"os"
	"fmt"
	"github.com/wengerbinning/srv/log"
	"github.com/wengerbinning/srv/config"
	"github.com/wengerbinning/srv/environment"
)

func main() {
	l, _ := log.CreateLog(nil, "")
	defer l.Close()

	conf, err := config.Load("config.yaml")
	if err != nil {
		l.Error( "Config error: %v", err)
		os.Exit(1)
	}

	sctx, err := environment.PrepareService(conf)
	if err != nil {
		l.Error( "Prepare error: %v", err)
		os.Exit(1)
	}
	defer sctx.Close()

	uctx := sctx.CreateUserContext()
	user, err := uctx.Registrate("alice", "alice@example.com", "password123")
	if err != nil {
		l.Error("Register error: %v", err)
		os.Exit(1)
	}
	l.Message("Registered user: id=%d username=%s email=%s", user.Uuid, user.Username, user.Email)

	updated, err := uctx.Changerate(user.Uuid, "alice_new", "alice_new@example.com", "newpassword")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update error: %v", err)
		os.Exit(1)
	}
	l.Message("Updated user: id=%d username=%s email=%s", updated.Uuid, updated.Username, updated.Email)

	if err := uctx.Cancellate(user.Uuid); err != nil {
		fmt.Fprintf(os.Stderr, "Delete error: %v", err)
		os.Exit(1)
	}
	l.Message( "Deleted user: id=%d", user.Uuid)
	l.Message( "Service started successfully")
}
