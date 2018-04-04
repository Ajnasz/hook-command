package aaa

import (
	"github.com/go-redis/redis"
	"io"
)

// RedisRangeReader reads lrange from redis
type RedisRangeReader struct {
	Client      *redis.Client
	Key         string
	rangeCursor int64
	len         int64
}

func (r RedisRangeReader) Read(p []byte) (n int, err error) {
	if r.len == 0 {
		maxLen, err := r.Client.LLen(r.Key).Result()

		if err != nil {
			return 0, err
		}

		if maxLen < 1 {
			return n, io.EOF
		}

		r.len = maxLen
	}

	if r.len <= r.rangeCursor {
		return n, io.EOF
	}

	for r.rangeCursor < r.len {
		values, err := r.Client.LRange(r.Key, r.rangeCursor, r.rangeCursor).Result()

		if err != nil {
			return 0, err
		}

		for _, value := range values {
			byteValue := []byte(value)

			if len(byteValue) > cap(p) {
				return n, err
			}

			for _, v := range byteValue {
				p[n] = v
				n++
			}
		}

		r.rangeCursor++

		if r.len <= r.rangeCursor {
			return n, io.EOF
		}
	}

	return n, err
}

// NewRedisRangeReader creates RedisRangeReader instance
func NewRedisRangeReader(client *redis.Client, key string) RedisRangeReader {
	return RedisRangeReader{
		Client: client,
		Key:    key,
	}
}
