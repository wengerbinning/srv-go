package environment

import (
	"fmt"
	"path/filepath"
	"github.com/wengerbinning/srv/config"
	"github.com/wengerbinning/srv/service"
	"github.com/wengerbinning/srv/database"
)

func PrepareService(conf *config.Conf) (*service.Context, error) {
	if err := ensureFold(conf.Storage.SrvPath); err != nil {
		return nil, fmt.Errorf("srv path: %w", err)
	}

	if err := ensureFold(conf.Storage.UsrPath); err != nil {
		return nil, fmt.Errorf("usr path: %w", err)
	}

	dbPath := filepath.Join(conf.Storage.SrvPath, conf.Storage.SrvDbName)
	sqllite, _ := database.SQLliteConf(dbPath)
	db, err := database.DataBaseInit(sqllite)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	return service.CreateContext(conf, db)
}

func ClearupService(ctx *service.Context) error {
	return service.DeleteContext(ctx)
}
