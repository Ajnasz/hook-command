package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

// RequestHandler Handles requests to the root path
func RequestHandler(w http.ResponseWriter, r *http.Request) {

	if err := testToken(r); err != nil {
		log.WithFields(err.LogFields).Error(err.LogFields)
		http.Error(w, err.Text, err.Code)
		return
	}

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

	go func() {
		errorLogger := logger{
			Fields: log.Fields{
				"step":  "runJob",
				"job":   jobName,
				"level": "Error",
			},
			LogLevel: logLevelInfo,
		}
		infoLogger := logger{
			Fields: log.Fields{
				"step":  "runJob",
				"job":   jobName,
				"level": "Info",
			},
			LogLevel: logLevelInfo,
		}
		for _, execConf := range execConfigs {

			jobEnd := make(chan int)

			outputs, err := runJob(execConf, w, jobEnd)

			if outputs == nil {
				log.WithFields(log.Fields{
					"step": "runJob",
					"job":  jobName,
				}).Error(err)
				// http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err != nil {
				log.WithFields(log.Fields{
					"step": "runJob",
					"job":  r.Header.Get(hookJobHeaderName),
				}).Error(err)

				// http.Error(w, "Unkown error", http.StatusInternalServerError)

				// w.Write([]byte("\n"))
			}

			writeProcessOutput(outputs, infoLogger, errorLogger)

			exitCode := <-jobEnd

			if exitCode != 0 {
				log.WithFields(log.Fields{
					"job":      r.Header.Get(hookJobHeaderName),
					"exitCode": exitCode,
				}).Error(errors.New("Job aborted"))
				break
			}
		}
	}()

	log.WithFields(log.Fields{
		"job": r.Header.Get(hookJobHeaderName),
	}).Info("Job accepted")

}
