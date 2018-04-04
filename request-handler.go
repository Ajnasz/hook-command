package main

import (
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	aaa "github.com/Ajnasz/hook-command/pkg"
	log "github.com/Sirupsen/logrus"
)

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(l int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	bytes := make([]byte, l)

	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(97, 122))
	}

	return string(bytes)
}

func execJob(jobName, redisKey string, execConfigs []ExecConf) {
	stdLogger := log.New()
	redisErrorLogger := log.New()
	redisErrorLogger.Out = NewRedisLogger(redisClient, redisKey+":error", log.ErrorLevel)

	redisInfoLogger := log.New()
	redisInfoLogger.Out = NewRedisLogger(redisClient, redisKey+":info", log.ErrorLevel)

	errorLogger := logger{
		Loggers: []io.Writer{
			logrusLogger{
				Fields: log.Fields{
					"step": "runJob",
					"job":  jobName,
				},
				LogLevel: log.ErrorLevel,
				Logger:   stdLogger,
			},
			logrusLogger{
				Fields: log.Fields{
					"step": "runJob",
					"job":  jobName,
				},
				LogLevel: log.ErrorLevel,
				Logger:   redisErrorLogger,
			},
		},
	}
	infoLogger := logger{
		Loggers: []io.Writer{
			logrusLogger{
				Fields: log.Fields{
					"step": "runJob",
					"job":  jobName,
				},
				LogLevel: log.InfoLevel,
				Logger:   stdLogger,
			},
			logrusLogger{
				Fields: log.Fields{
					"step": "runJob",
					"job":  jobName,
				},
				LogLevel: log.InfoLevel,
				Logger:   redisInfoLogger,
			},
		},
	}
	for _, execConf := range execConfigs {

		jobEnd := make(chan int)

		outputs, err := runJob(execConf, jobEnd)

		if outputs == nil {
			errorLogger.Write([]byte(err.Error()))
			break
		}

		if err != nil {
			errorLogger.Write([]byte(err.Error()))
		}

		writeProcessOutput(outputs, infoLogger, errorLogger)

		exitCode := <-jobEnd

		if exitCode != 0 {
			errorLogger.Write([]byte("Job exited with code " + strconv.Itoa(exitCode)))
			break
		} else {
			infoLogger.Write([]byte("Job exited with code " + strconv.Itoa(exitCode)))
		}
	}

	infoLogger.Write([]byte("EOL"))
}

func handleNewJobRequest(w http.ResponseWriter, r *http.Request) {
	jobName := r.Header.Get(hookJobHeaderName)

	execConfigs, err := getExecConfigs(r)

	if err != nil {
		log.WithFields(log.Fields{
			"job": jobName,
		}).Error(err)

		http.Error(w, "Unknown error", http.StatusInternalServerError)

		return
	}

	if !hasConfigs(execConfigs) {
		log.WithFields(log.Fields{
			"job": jobName,
		}).Error("Configuration not found")
		http.NotFound(w, r)

		return
	}

	execConfigs, err = extendExecConfigs(r, execConfigs)

	if err != nil {
		http.Error(w, "BadRequest", http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{
		"job": jobName,
	}).Info("Job start")

	redisKey := randomString(16)
	w.Write([]byte(redisKey))

	go execJob(jobName, redisKey, execConfigs)

	log.WithFields(log.Fields{
		"job": r.Header.Get(hookJobHeaderName),
	}).Info("Job accepted")

}

type jobResponse struct {
	Info  []string `json:"info"`
	Error []string `json:"error"`
}

func handleGetJob(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.SplitAfter(r.URL.Path, "/job/")

	jobID := pathSplit[1]

	infos := aaa.NewRedisRangeReader(redisClient, "redis_logs:"+jobID+":info")
	errors := aaa.NewRedisRangeReader(redisClient, "redis_logs:"+jobID+":error")
	io.Copy(w, infos)
	io.Copy(w, errors)
}

// RequestHandler Handles requests to the root path
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	if err := testToken(r); err != nil {
		log.WithFields(err.LogFields).Error(err.LogFields)
		http.Error(w, err.Text, err.Code)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/job/") {
		handleGetJob(w, r)
		return
	}

	handleNewJobRequest(w, r)
}
