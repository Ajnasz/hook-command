package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
)

const logLevelError string = "error"
const logLevelInfo string = "info"

type logger struct {
	Fields   log.Fields
	LogLevel string
}

func (l logger) Write(p []byte) (n int, err error) {
	if l.LogLevel == logLevelError {
		log.WithFields(l.Fields).Error(string(p))

		return len(p), nil
	} else if l.LogLevel == logLevelInfo {
		log.WithFields(l.Fields).Info(string(p))
		return len(p), nil
	}

	return 0, errors.New("No such loglevel")
}
