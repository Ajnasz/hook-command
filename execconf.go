package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ExecConf is a struct to define configuration of execution configs
type ExecConf struct {
	Job     string   `json:"job"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env"`
	Dir     string   `json:"dir"`
}

func readExecConfFile(filename string) ([]ExecConf, error) {
	configFilePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	file, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		return nil, err
	}

	var execConfigs []ExecConf

	json.Unmarshal(file, &execConfigs)

	return execConfigs, nil
}

func readExecConfDir(directory string) ([]ExecConf, error) {
	configDirPath, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}

	var execConfigs []ExecConf
	err = filepath.Walk(configDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileConfigs, err := readExecConfFile(path)

		if err != nil {
			return err
		}

		execConfigs = append(execConfigs, fileConfigs...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return execConfigs, nil
}
