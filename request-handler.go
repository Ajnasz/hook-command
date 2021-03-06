package main

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Ajnasz/hook-command/execjob"
	"github.com/Ajnasz/hook-command/redisrangereader"
	log "github.com/Sirupsen/logrus"
)

const redisKeyPrefix string = "redis_logs:"

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

	msg, err := json.Marshal(execjob.Job{
		JobName:     jobName,
		RedisKey:    redisKey,
		ExecConfigs: execConfigs,
	})

	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	redisClient.Publish(jobChannelName, msg)

	log.WithFields(log.Fields{
		"job": r.Header.Get(hookJobHeaderName),
	}).Info("Job accepted")

}

type jobResponse struct {
	Info  []string `json:"info"`
	Error []string `json:"error"`
}

func hasJob(jobID string) bool {
	keys, err := redisClient.Keys(redisKeyPrefix + jobID).Result()

	return err == nil && len(keys) == 1
}

func handleGetJob(w http.ResponseWriter, r *http.Request) {
	pathSplit := strings.SplitAfter(r.URL.Path, "/job/")

	jobID := pathSplit[1]

	if !hasJob(jobID) {
		http.NotFound(w, r)
		return
	}

	infos := redisrangereader.NewRedisRangeReader(redisClient, redisKeyPrefix+jobID)
	io.Copy(w, infos)
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
