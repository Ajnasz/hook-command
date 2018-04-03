package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
)

const logLevelError string = "error"
const logLevelInfo string = "info"

type logrusLogger struct {
	Fields   log.Fields
	LogLevel string
}

func (l logrusLogger) Write(p []byte) (n int, err error) {
	if l.LogLevel == logLevelError {
		log.WithFields(l.Fields).Error(string(p))
		return len(p), nil
	} else if l.LogLevel == logLevelInfo {
		log.WithFields(l.Fields).Info(string(p))
		return len(p), nil
	}

	return 0, errors.New("No such loglevel")
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
