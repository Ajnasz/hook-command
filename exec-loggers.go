package main

import (
	"io"

	"github.com/Ajnasz/hook-command/logger"
	log "github.com/Sirupsen/logrus"
)

type execLoggers struct {
	Info  logger.Logger
	Error logger.Logger
}

func getExecLoggers(jobName string, stdLogger *log.Logger) execLoggers {
	errorLogger := logger.Logger{
		Loggers: []io.Writer{
			logger.LogrusLogger{
				Fields: log.Fields{
					"job": jobName,
				},
				LogLevel: log.ErrorLevel,
				Logger:   stdLogger,
			},
		},
	}

	infoLogger := logger.Logger{
		Loggers: []io.Writer{
			logger.LogrusLogger{
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
