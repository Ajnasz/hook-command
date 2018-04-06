package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/kelseyhightower/envconfig"
)

var config Config
var redisClient *redis.Client

const hookTokenHeaderName string = "X-HOOK-TOKEN"
const hookJobHeaderName string = "X-HOOK-JOB"

// ErrNoExecConf an error which used when no execution configuration foiund for a key
var ErrNoExecConf = errors.New("No execConf found")

func hasValidToken(r *http.Request) bool {
	return r.Header.Get(hookTokenHeaderName) == config.Token
}

func getExecConfigs(r *http.Request) ([]ExecConf, error) {
	job := r.Header.Get(hookJobHeaderName)

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

func runJob(execConf ExecConf, finish chan int) (*ProcessOutput, error) {
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

func writeProcessOutput(outputs *ProcessOutput, info logger, err logger) {

	outputChan := make(chan []byte)
	errorChan := make(chan []byte)
	q := make(chan int)

	go scanIO(outputs.Stdout, outputChan, q)
	go scanIO(outputs.Stderr, errorChan, q)

	quitCount := 0
	for {
		select {
		case errBytes := <-errorChan:
			err.Write(errBytes)
		case outBytes := <-outputChan:
			info.Write(outBytes)
		case <-q:
			quitCount++
		}

		if quitCount >= 2 {
			break
		}
	}
}

func extendExecConfigs(r *http.Request, execConfigs []ExecConf) ([]ExecConf, error) {
	body, err := getJSONBody(r)

	if err != nil {
		return nil, err
	}

	for i, execConfig := range execConfigs {
		execConfigs[i] = extendExecConfig(execConfig, body)
	}

	return execConfigs, nil
}

func testToken(r *http.Request) *MiddlewareError {
	if !hasValidToken(r) {
		return &MiddlewareError{
			http.StatusForbidden,
			"Forbidden",
			log.Fields{
				"job": r.Header.Get(hookJobHeaderName),
			},
			"Invalid token",
		}
	}

	return nil
}

func init() {
	err := envconfig.Process("HCMD", &config)

	if err != nil {
		log.Fatal(err)
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
}

func main() {
	http.HandleFunc("/", RequestHandler)

	log.Info(fmt.Sprintf("Listening on port %d", config.Port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
