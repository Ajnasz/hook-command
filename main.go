package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
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

func hasValidToken(r *http.Request) bool {
	return r.Header.Get("X-HOOK-TOKEN") == config.Token
}

func getExecConfigs(r *http.Request) ([]ExecConf, error) {
	job := r.Header.Get("X-HOOK-JOB")

	if job == "" {
		return nil, nil
	}

	file, err := ioutil.ReadFile(config.ConfigFile)

	if err != nil {
		log.Error("Configuration file not found")
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

// ProcessOutput a struct to store process std out and std err
type ProcessOutput struct {
	Stdout []byte
	Stderr []byte
}

func runJob(execConf ExecConf, w http.ResponseWriter) (*ProcessOutput, error) {
	cmd := exec.Cmd{
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

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	stdoutBytes, err := ioutil.ReadAll(stdout)

	if err != nil {
		return nil, err
	}

	stderrBytes, err := ioutil.ReadAll(stderr)

	if err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return &ProcessOutput{stdoutBytes, stderrBytes}, err
	}

	return &ProcessOutput{stdoutBytes, stderrBytes}, nil
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

	if len(execConfigs) == 0 {
		log.WithFields(log.Fields{
			"job": r.Header.Get("X-HOOK-JOB"),
		}).Error("Configuration not found")
		http.NotFound(w, r)
		return
	}

	for _, execConf := range execConfigs {

		outputs, err := runJob(execConf, w)

		if outputs == nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if runJobErr, ok := err.(*exec.ExitError); ok {
			log.WithFields(log.Fields{
				"step":   "runJob",
				"job":    r.Header.Get("X-HOOK-JOB"),
				"stdout": string(outputs.Stdout),
				"stderr": string(outputs.Stderr),
			}).Error(runJobErr)

			http.Error(w, runJobErr.Error(), http.StatusInternalServerError)

			w.Write([]byte("\n"))
			return
		}

		if outputs.Stdout != nil {
			stdOutScanner := bufio.NewScanner(bytes.NewReader(outputs.Stdout))
			for stdOutScanner.Scan() {
				w.Write([]byte("OUT: "))
				w.Write(stdOutScanner.Bytes())
				w.Write([]byte("\n"))
			}
		}

		if outputs.Stderr != nil {
			stdErrScanner := bufio.NewScanner(bytes.NewReader(outputs.Stderr))
			for stdErrScanner.Scan() {
				w.Write([]byte("ERR: "))
				w.Write(stdErrScanner.Bytes())
				w.Write([]byte("\n"))
			}
		}
	}
	log.WithFields(log.Fields{
		"job": r.Header.Get("X-HOOK-JOB"),
	}).Info("Job run")

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
