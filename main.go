package main

import (
	"bufio"
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

// Config is a struct to define configuration
type Config struct {
	Port       int    `default:"10292"`
	ConfigFile string `default:"./configuration.json"`
	ScriptsDir string `default:"./scripts"`
	Token      string `required:"true"`
}

// ExecConf is a struct to define configuration of execution configs
type ExecConf struct {
	Job     string   `json:"job"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env"`
	Dir     string   `json:"dir"`
}

var config Config

// ErrNoExecConf an error which used when no execution configuration foiund for a key
var ErrNoExecConf = errors.New("No execConf found")

// ProcessOutput a struct to store process std out and std err
type ProcessOutput struct {
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

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

func runJob(execConf ExecConf, w http.ResponseWriter) (*ProcessOutput, error) {
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

	log.WithFields(log.Fields{
		"job": r.Header.Get("X-HOOK-JOB"),
	}).Info("Job start")

	for _, execConf := range execConfigs {

		outputs, err := runJob(execConf, w)

		if outputs == nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if runJobErr, ok := err.(*exec.ExitError); ok {
			log.WithFields(log.Fields{
				"step": "runJob",
				"job":  r.Header.Get("X-HOOK-JOB"),
			}).Error(runJobErr)

			http.Error(w, runJobErr.Error(), http.StatusInternalServerError)

			w.Write([]byte("\n"))
		}

		writeProcessOutput(outputs, w)
	}

	log.WithFields(log.Fields{
		"job": r.Header.Get("X-HOOK-JOB"),
	}).Info("Job finished")

}

func init() {
	err := envconfig.Process("scm", &config)

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", HomeHandler)

	log.Info(fmt.Sprintf("Listening on port %d", config.Port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
