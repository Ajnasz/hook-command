package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
)

// RedisLogger logs things to redis
type RedisLogger struct {
	client *redis.Client
	key    string
}

func (l RedisLogger) Write(b []byte) (n int, err error) {
	l.client.RPush(l.key, b)
	l.client.Expire(l.key, time.Duration(1)*time.Hour)

	return len(b), nil
}

// NewRedisLogger creates new redis logger
func NewRedisLogger(client *redis.Client, key string, logLevel log.Level) *RedisLogger {
	return &RedisLogger{
		client: client,
		key:    key,
	}
}
