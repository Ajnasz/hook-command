package main

// Config is a struct to define configuration
type Config struct {
	Port          int    `default:"10292"`
	ConfigFile    string `required:"true"`
	ScriptsDir    string `required:"true"`
	Token         string `required:"true"`
	RedisAddr     string `default:"localhost:6379"`
	RedisPassword string
	RedisDB       int `default:"1"`
}
