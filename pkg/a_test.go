package aaa

import (
	"github.com/go-redis/redis"
	"io/ioutil"
	"testing"
)

func createClient(keys []string) (redisClient *redis.Client, err error) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       2,
	})

	for _, k := range keys {
		err := redisClient.Del(k).Err()

		if err != nil {
			return nil, err
		}
	}

	return
}

func TestReaderEmpty(t *testing.T) {
	key := "foo:bar"
	redisClient, err := createClient([]string{key})

	if err != nil {
		t.Error(err)
	}

	reader := NewRedisRangeReader(redisClient, key)
	stuff, err := ioutil.ReadAll(reader)

	if err != nil {
		t.Error(err)
	}

	if len(stuff) > 0 {
		t.Error("Unexpected returned value", stuff)
	}

	redisClient.Close()
}

func TestReaderNotEmpty(t *testing.T) {
	key := "foo:bar"
	value := "asdfsdflkadjsf"

	redisClient, err := createClient([]string{key})

	if err != nil {
		t.Error(err)
	}

	err = redisClient.Del(key).Err()

	if err != nil {
		t.Error(err)
	}

	err = redisClient.RPush(key, value).Err()

	if err != nil {
		t.Error(err)
	}

	reader := NewRedisRangeReader(redisClient, key)
	stuff, err := ioutil.ReadAll(reader)

	if err != nil {
		t.Error(err)
	}

	if string(stuff) != value {
		t.Error(string(stuff), "is not as expected", value)
	}

	redisClient.Close()
}
