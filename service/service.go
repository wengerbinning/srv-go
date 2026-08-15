package service

import (
	"database/sql"
	"github.com/wengerbinning/srv/log"
	"github.com/wengerbinning/srv/config"
)

type Context struct {
	conf *config.Conf
	db   *sql.DB
	log  log.Interface
}

func logConf(conf *config.Conf) *log.Conf {
	return &log.Conf{
		Stdio:   conf.Log.Stdio,
		File:    conf.Log.File,
		FilePath: conf.Log.FilePath,
		Syslog:  conf.Log.Syslog,
		Systag:  conf.Log.Systag,
	}
}

func CreateContext(conf *config.Conf, db *sql.DB) (*Context, error) {
	logger, err := log.CreateLog(logConf(conf), "srvlog")
	if err != nil {
		return nil, err
	}

	return &Context{conf: conf, db: db, log: logger}, nil
}

func DeleteContext(s *Context) error {
	return s.log.Close()
}

func (s *Context) Log(level log.Level, format string, v ...any) {
	switch level {
	case log.Error:
		s.log.Error(format, v...)
	case log.Warning:
		s.log.Warning(format, v...)
	case log.Message:
		s.log.Message(format, v...)
	case log.Notice:
		s.log.Notice(format, v...)
	case log.Debug:
		s.log.Debug(format, v...)
	}
}

func (s *Context) Close() error {
	return DeleteContext(s)
}
