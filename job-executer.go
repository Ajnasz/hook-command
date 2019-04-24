package main

import (
	"strconv"

	"github.com/Ajnasz/logrus-redis"
	log "github.com/Sirupsen/logrus"
)

func execJob(jobName, redisKey string, execConfigs []ExecConf) {
	stdLogger := log.New()
	hook := logrusredis.NewLogrusRedis(redisClient, redisKeyPrefix+redisKey)

	stdLogger.Hooks.Add(hook)

	loggers := getExecLoggers(jobName, stdLogger)

	for _, execConf := range execConfigs {

		jobEnd := make(chan int)

		outputs, err := runJob(execConf, jobEnd)

		if outputs == nil {
			loggers.Error.Write([]byte(err.Error()))
			break
		}

		if err != nil {
			loggers.Error.Write([]byte(err.Error()))
		}

		writeProcessOutput(outputs, loggers)

		exitCode := <-jobEnd

		if exitCode != 0 {
			loggers.Error.Write([]byte("Job exited with code " + strconv.Itoa(exitCode)))
			break
		} else {
			loggers.Info.Write([]byte("Job exited with code " + strconv.Itoa(exitCode)))
		}
	}

	loggers.Info.Write([]byte("EOL"))
}
