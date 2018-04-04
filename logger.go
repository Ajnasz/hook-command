package main

import (
	log "github.com/Sirupsen/logrus"
	"io"
)

type logrusLogger struct {
	Fields   log.Fields
	LogLevel log.Level
	Logger   *log.Logger
}

func (l logrusLogger) Write(p []byte) (n int, err error) {
	entry := l.Logger.WithFields(l.Fields)

	switch l.LogLevel {
	case log.ErrorLevel:
		entry.Error(string(p))
		return len(p), nil
	case log.InfoLevel:
		entry.Info(string(p))
		return len(p), nil
	default:
		return 0, nil
	}
}

type logger struct {
	Loggers []io.Writer
}

func (l logger) Write(p []byte) (n int, err error) {
	for _, l := range l.Loggers {
		_, err := l.Write(p)

		if err != nil {
			return 0, err
		}
	}

	return len(p), nil
}
