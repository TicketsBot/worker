package redis

import (
	"github.com/go-redis/redis"
	"os"
	"strconv"
)

var Client *redis.Client

func Connect() error {
	threads, err := strconv.Atoi(os.Getenv("WORKER_REDIS_THREADS"))
	if err != nil {
		return err
	}

	Client = redis.NewClient(&redis.Options{
		Network:            "tcp",
		Addr:               os.Getenv("WORKER_REDIS_ADDR"),
		Password:           os.Getenv("WORKER_REDIS_PASSWD"),
		PoolSize:           threads,
		MinIdleConns:       threads,
	})

	return nil
}
