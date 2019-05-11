package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Ajnasz/hook-command/execconf"
	log "github.com/Sirupsen/logrus"
)

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

func extendExecConfig(execConfig execconf.ExecConf, jsonBody *JSONBody) execconf.ExecConf {
	for name, value := range jsonBody.Env {
		execConfig.Env = append(execConfig.Env, fmt.Sprintf("%s=%s", name, value))
	}

	return execConfig
}
