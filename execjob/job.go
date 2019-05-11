package execjob

import "github.com/Ajnasz/hook-command/execconf"

// Job a struct to represent a job in json
type Job struct {
	JobName     string              `json:"jobName"`
	RedisKey    string              `json:"redisKey"`
	ExecConfigs []execconf.ExecConf `json:"execConfigs"`
}
