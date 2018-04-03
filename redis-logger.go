package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"time"
)

// RedisLogger logs things to redis
type RedisLogger struct {
	client   *redis.Client
	key      string
	Fields   log.Fields
	LogLevel log.Level
}

func (l RedisLogger) Write(b []byte) (n int, err error) {
	formatter := log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	}

	entry := log.WithFields(l.Fields)
	entry.Message = string(b)
	entry.Time = time.Now()
	entry.Level = l.LogLevel

	str, err := formatter.Format(entry)

	if err != nil {
		return 0, err
	}

	l.client.RPush(l.key, str)
	l.client.Expire(l.key, time.Duration(1)*time.Hour)

	return len(b), nil
}

// Get returns log result
func (l RedisLogger) Get(subGroup string) ([]string, error) {

	var output = []string{}

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = l.client.Scan(cursor, l.key+":"+subGroup, 10).Result()
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
func NewRedisLogger(client *redis.Client, key string, fields log.Fields, logLevel log.Level) *RedisLogger {
	return &RedisLogger{
		client,
		"redis_logs:" + key,
		fields,
		logLevel,
	}
}
