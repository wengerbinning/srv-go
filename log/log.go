package log

import (
	"io"
	"os"
	"fmt"
	"time"
	"log/syslog"
	"path/filepath"
)

type Level int

const (
	Error   Level = iota
	Warning
	Message
	Notice
	Debug
)

var levelNames = [5]string{
	"error",
	"warning",
	"message",
	"notice",
	"debug",
}

func (p Level) String() string {
	return levelNames[p]
}

type Conf struct {
	Stdio      bool
	File       bool
	FilePath string
	Syslog     bool
	Systag   string
}

type Log struct {
	name    string
	conf    Conf
	stdio   io.Writer
	file    io.Writer
	syslog  *syslog.Writer
	closers []io.Closer
}

type Interface interface {
	Error(format string, v ...any)
	Warning(format string, v ...any)
	Message(format string, v ...any)
	Notice(format string, v ...any)
	Debug(format string, v ...any)
	Close() error
}

func (l *Log) logFile(prio Level, format string, v ...any) {
	if l.file == nil {
		return
	}
	fmt.Fprintf(l.file, time.Now().Format("2006-01-02 15:04:05 ") +
		prio.String()+ ": " + format + "\n", v...)
}

func (l *Log) logStdio(prio Level, format string, v ...any) {
	if l.stdio == nil {
		return
	}
	if l.name != "" {
		fmt.Fprintf(l.stdio, l.name+"."+prio.String()+": "+format+"\n", v...)
	} else {
		fmt.Fprintf(l.stdio, prio.String()+": "+format+"\n", v...)
	}
}

func (l *Log) logSyslog(prio Level, format string, v ...any) {
	if l.syslog == nil {
		return
	}
	msg := fmt.Sprintf(format, v...)
	switch prio {
	case Error:
		l.syslog.Err(msg)
	case Warning:
		l.syslog.Warning(msg)
	case Message:
		l.syslog.Info(msg)
	case Notice:
		l.syslog.Notice(msg)
	case Debug:
		l.syslog.Debug(msg)
	}
}

func (l *Log) Error(format string, v ...any) {
	l.logFile(Error, format, v...)
	l.logStdio(Error, format, v...)
	l.logSyslog(Error, format, v...)
}

func (l *Log) Warning(format string, v ...any) {
	l.logFile(Warning, format, v...)
	l.logStdio(Warning, format, v...)
	l.logSyslog(Warning, format, v...)
}

func (l *Log) Message(format string, v ...any) {
	l.logFile(Message, format, v...)
	l.logStdio(Message, format, v...)
	l.logSyslog(Message, format, v...)
}

func (l *Log) Notice(format string, v ...any) {
	l.logFile(Notice, format, v...)
	l.logStdio(Notice, format, v...)
	l.logSyslog(Notice, format, v...)
}

func (l *Log) Debug(format string, v ...any) {
	l.logFile(Debug, format, v...)
	l.logStdio(Debug, format, v...)
	l.logSyslog(Debug, format, v...)
}

func (l *Log) Close() error {
	for _, c := range l.closers {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}

func CreateLog(conf *Conf, name string) (Interface, error) {
	var closers []io.Closer

	if conf == nil {
		return &Log{name: name,
			stdio: os.Stderr, file: nil, syslog: nil, closers: closers}, nil
	}

	var stdioWriter, fileWriter io.Writer
	if conf.Stdio {
		stdioWriter = os.Stdout
	}
	if conf.File {
		file := name
		if file == "" {
			file = "srvlog"
		}
		fold := conf.FilePath
		if err := os.MkdirAll(fold, 0755); err != nil {
			return nil, fmt.Errorf("create log dir: %w", err)
		}
		path := filepath.Join(fold, file +".log")
		w, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		fileWriter = w
		closers = append(closers, w)
	}

	var syslogWriter *syslog.Writer
	if conf.Syslog {
		w, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, conf.Systag)
		if err != nil {
			for _, c := range closers {
				c.Close()
			}
			return nil, fmt.Errorf("open syslog: %w", err)
		}
		syslogWriter = w
		closers = append(closers, w)
	}

	return &Log{conf: *conf, name: name,
		stdio: stdioWriter, file: fileWriter, syslog: syslogWriter,
		closers: closers}, nil
}

func DeleteLog(lg Interface) error {
	return lg.Close()
}
