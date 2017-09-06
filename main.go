package main

import (
	"bufio"
	"syscall"
	// "bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
)

var config Config

// ErrNoExecConf an error which used when no execution configuration foiund for a key
var ErrNoExecConf = errors.New("No execConf found")

func hasValidToken(r *http.Request) bool {
	return r.Header.Get("X-HOOK-TOKEN") == config.Token
}

func getExecConfigs(r *http.Request) ([]ExecConf, error) {
	job := r.Header.Get("X-HOOK-JOB")

	if job == "" {
		return nil, nil
	}

	configFilePath, err := filepath.Abs(config.ConfigFile)
	if err != nil {
		log.WithFields(log.Fields{
			"configFilePath": configFilePath,
		}).Error(err)
		return nil, err
	}

	file, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		log.WithFields(log.Fields{
			"configFilePath": configFilePath,
		}).Error(err)
		return nil, err
	}

	var execConfigs []ExecConf

	json.Unmarshal(file, &execConfigs)

	var output []ExecConf

	for _, execConf := range execConfigs {
		if execConf.Job == job {
			output = append(output, execConf)
		}
	}

	return output, nil
}

func getCmd(execConf ExecConf) (*exec.Cmd, error) {
	cmd := &exec.Cmd{
		Path: filepath.Join(execConf.Command),
	}

	if execConf.Args != nil && len(execConf.Args) > 0 {
		cmd.Args = execConf.Args
	}

	if execConf.Env != nil && len(execConf.Env) > 0 {
		cmd.Env = execConf.Env
	}

	if execConf.Dir == "" {
		absPath, err := filepath.Abs(config.ScriptsDir)

		if err != nil {
			return nil, err
		}

		cmd.Dir = absPath
	} else {
		cmd.Dir = filepath.Join(config.ScriptsDir, execConf.Dir)
	}

	return cmd, nil
}

func runJob(execConf ExecConf, w http.ResponseWriter, finish chan int) (*ProcessOutput, error) {
	cmd, err := getCmd(execConf)

	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		err := cmd.Wait()

		if err == nil {
			finish <- 0
		} else if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			finish <- waitStatus.ExitStatus()
		}
	}()

	return &ProcessOutput{stdout, stderr}, nil
}

func scanIO(io io.Reader, c chan []byte, done chan int) {
	scanner := bufio.NewScanner(io)
	for scanner.Scan() {
		c <- scanner.Bytes()
	}

	done <- 1
}

func hasConfigs(execConfigs []ExecConf) bool {
	return len(execConfigs) > 0
}

func writeProcessOutput(outputs *ProcessOutput, w http.ResponseWriter) {
	flusher := w.(http.Flusher)

	oChan := make(chan []byte)
	eChan := make(chan []byte)
	q := make(chan int)

	go scanIO(outputs.Stdout, oChan, q)
	go scanIO(outputs.Stderr, eChan, q)

	qnum := 0
	for {
		select {
		case errBytes := <-eChan:
			w.Write([]byte("ERR: "))
			w.Write(errBytes)
			w.Write([]byte("\n"))
			flusher.Flush()
		case outBytes := <-oChan:
			w.Write([]byte("OUT: "))
			w.Write(outBytes)
			w.Write([]byte("\n"))
			flusher.Flush()
		case <-q:
			qnum++
		}

		if qnum >= 2 {
			break
		}
	}
}

// JSONBody defines the json format of request body
type JSONBody struct {
	Env map[string]string `json:"ENV"`
}

func getJSONBody(r *http.Request) (*JSONBody, error) {
	decoder := json.NewDecoder(r.Body)

	var output JSONBody

	err := decoder.Decode(&output)

	if err != nil {
		log.WithError(err).Error("Could not parse request body")
		return nil, err
	}

	defer r.Body.Close()

	return &output, nil
}

func extendExecConfig(execConfig ExecConf, jsonBody *JSONBody) ExecConf {
	for name, value := range jsonBody.Env {
		execConfig.Env = append(execConfig.Env, fmt.Sprintf("%s=%s", name, value))
	}

	return execConfig
}

func extendExecConfigs(r *http.Request, execConfigs []ExecConf) []ExecConf {
	body, err := getJSONBody(r)

	if err != nil {
		return execConfigs
	}

	for i, execConfig := range execConfigs {
		execConfigs[i] = extendExecConfig(execConfig, body)
	}

	return execConfigs
}

// HomeHandler Handles requests to the root path
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	if !hasValidToken(r) {
		log.WithFields(log.Fields{
			"job": r.Header.Get("X-HOOK-JOB"),
		}).Error("Invalid token")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	execConfigs, err := getExecConfigs(r)

	if err != nil {
		log.WithFields(log.Fields{
			"job": r.Header.Get("X-HOOK-JOB"),
		}).Error(err)

		http.Error(w, "Unknown error", http.StatusInternalServerError)

		return
	}

	if !hasConfigs(execConfigs) {
		log.WithFields(log.Fields{
			"job": r.Header.Get("X-HOOK-JOB"),
		}).Error("Configuration not found")
		http.NotFound(w, r)
		return
	}

	execConfigs = extendExecConfigs(r, execConfigs)

	log.WithFields(log.Fields{
		"job": r.Header.Get("X-HOOK-JOB"),
	}).Info("Job start")

	for _, execConf := range execConfigs {

		jobEnd := make(chan int)

		outputs, err := runJob(execConf, w, jobEnd)

		if outputs == nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err != nil {
			log.WithFields(log.Fields{
				"step": "runJob",
				"job":  r.Header.Get("X-HOOK-JOB"),
			}).Error(err)

			http.Error(w, "Unkown error", http.StatusInternalServerError)

			w.Write([]byte("\n"))
		}

		writeProcessOutput(outputs, w)

		exitCode := <-jobEnd

		if exitCode != 0 {
			log.WithFields(log.Fields{
				"job":      r.Header.Get("X-HOOK-JOB"),
				"exitCode": exitCode,
			}).Error(errors.New("Job aborted"))
			break
		}
	}

	log.WithFields(log.Fields{
		"job": r.Header.Get("X-HOOK-JOB"),
	}).Info("Job finished")

}

func init() {
	err := envconfig.Process("HCMD", &config)

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", HomeHandler)

	log.Info(fmt.Sprintf("Listening on port %d", config.Port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
