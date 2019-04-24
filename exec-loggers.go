package main

import (
	"io"

	log "github.com/Sirupsen/logrus"
)

type execLoggers struct {
	Info  logger
	Error logger
}

func getExecLoggers(jobName string, stdLogger *log.Logger) execLoggers {
	errorLogger := logger{
		Loggers: []io.Writer{
			logrusLogger{
				Fields: log.Fields{
					"job": jobName,
				},
				LogLevel: log.ErrorLevel,
				Logger:   stdLogger,
			},
		},
	}

	infoLogger := logger{
		Loggers: []io.Writer{
			logrusLogger{
				Fields: log.Fields{
					"job": jobName,
				},
				LogLevel: log.InfoLevel,
				Logger:   stdLogger,
			},
		},
	}

	return execLoggers{infoLogger, errorLogger}
}
