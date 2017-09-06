package main

// ExecConf is a struct to define configuration of execution configs
type ExecConf struct {
	Job     string   `json:"job"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env"`
	Dir     string   `json:"dir"`
}
