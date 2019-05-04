package main

import (
	"fmt"
	"strconv"

	"github.com/Ajnasz/logrus-redis"
	log "github.com/Sirupsen/logrus"
)

func execJob(jobName, redisKey string, execConfigs []ExecConf) {
	stdLogger := log.New()
	hook := logrusredis.NewLogrusRedis(redisClient, redisKeyPrefix+redisKey)

	stdLogger.Hooks.Add(hook)

	loggers := getExecLoggers(jobName, stdLogger)
	loggers.Info.Write([]byte("execute job package"))

	for _, execConf := range execConfigs {
		loggers.Info.Write([]byte(fmt.Sprintf("execute job %s", execConf.Command)))

		jobEnd := make(chan int)

		outputs, err := runJob(execConf, jobEnd)

		if outputs == nil {
			loggers.Error.Write([]byte(err.Error()))
			break
		}

		if err != nil {
			loggers.Error.Write([]byte(err.Error()))
			loggers.Error.Write([]byte("Job exection failed"))
			return
		}

		writeProcessOutput(outputs, loggers)

		exitCode := <-jobEnd

		if exitCode != 0 {
			loggers.Error.Write([]byte("Job exited with code " + strconv.Itoa(exitCode)))
			break
		} else {
			loggers.Info.Write([]byte("Job exited with code " + strconv.Itoa(exitCode)))
		}

		loggers.Info.Write([]byte(fmt.Sprintf("execute job finished %s", execConf.Command)))
	}

	loggers.Info.Write([]byte("execute job package finished"))
	loggers.Info.Write([]byte("EOL"))
}
