package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
)

// RedisLogger logs things to redis
type RedisLogger struct {
	client     *redis.Client
	key        string
	LogLevel   log.Level
	keysCursor uint64
}

func (l RedisLogger) Write(b []byte) (n int, err error) {
	l.client.RPush(l.key, b)
	l.client.Expire(l.key, time.Duration(1)*time.Hour)

	return len(b), nil
}

// Get returns log result
func (l RedisLogger) Get() ([]string, error) {

	var output = []string{}

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = l.client.Scan(cursor, l.key, 10).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			values := l.client.LRange(key, 0, -1).Val()
			output = append(output, values...)
		}

		if cursor == 0 {
			break
		}
	}

	return output, nil
}

// NewRedisLogger creates new redis logger
func NewRedisLogger(client *redis.Client, key string, logLevel log.Level) *RedisLogger {
	return &RedisLogger{
		client:   client,
		key:      "redis_logs:" + key,
		LogLevel: logLevel,
	}
}
