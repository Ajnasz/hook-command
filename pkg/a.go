package aaa

import (
	"github.com/go-redis/redis"
	"io"
)

// RedisRangeReader reads lrange from redis
type RedisRangeReader struct {
	client      *redis.Client
	key         string
	rangeCursor int64
	len         int64
}

func (r RedisRangeReader) Read(p []byte) (n int, err error) {
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

	for r.rangeCursor < r.len {
		values, err := r.client.LRange(r.key, r.rangeCursor, r.rangeCursor+10).Result()

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

			r.rangeCursor++
		}

		if r.len <= r.rangeCursor {
			return n, io.EOF
		}
	}

	return n, err
}

// NewRedisRangeReader creates RedisRangeReader instance
func NewRedisRangeReader(client *redis.Client, key string) RedisRangeReader {
	return RedisRangeReader{
		client: client,
		key:    key,
	}
}
