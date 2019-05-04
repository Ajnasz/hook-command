package redisrangereader

import (
	"errors"
	"io"

	"github.com/go-redis/redis"
)

// RedisRangeReader reads lrange from redis
type RedisRangeReader struct {
	client       *redis.Client
	key          string
	rangeCursor  int64
	LinesPerRead int64
	len          int64
}

func (r *RedisRangeReader) Read(p []byte) (n int, err error) {
	if r.LinesPerRead < 1 {
		return 0, errors.New("LinesPerRead must be greater then 0")
	}
	if r.len == 0 {
		maxLen, err := r.client.LLen(r.key).Result()

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

	pCap := cap(p)

	for r.rangeCursor < r.len {
		values, err := r.client.LRange(r.key, r.rangeCursor, r.rangeCursor+r.LinesPerRead).Result()

		if err != nil {
			return 0, err
		}

		for _, value := range values {
			byteValue := []byte(value)

			if n+len(byteValue) >= pCap {
				return n, err
			}

			for _, v := range byteValue {
				p[n] = v
				n++
			}
			r.rangeCursor++
		}

		if r.len <= r.rangeCursor {
			return n, io.EOF
		}
	}

	return n, err
}

// NewRedisRangeReader creates RedisRangeReader instance
func NewRedisRangeReader(client *redis.Client, key string) *RedisRangeReader {
	return &RedisRangeReader{
		client:       client,
		key:          key,
		LinesPerRead: 100,
	}
}
