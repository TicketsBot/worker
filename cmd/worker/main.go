package main

import (
	"github.com/TicketsBot/worker/event"
	"github.com/go-redis/redis"
	"os"
	"strconv"
)

func main() {
	// create redis client
	redis, err := buildRedisClient()
	if err != nil {
		panic(err)
	}

	event.Listen(redis)
}

func buildRedisClient() (client *redis.Client, err error) {
	threads, err := strconv.Atoi(os.Getenv("WORKER_REDIS_THREADS"))
	if err != nil {
		return
	}

	client = redis.NewClient(&redis.Options{
		Network:            "tcp",
		Addr:               os.Getenv("WORKER_REDIS_ADDR"),
		Password:           os.Getenv("WORKER_REDIS_PASSWD"),
		PoolSize:           threads,
		MinIdleConns:       threads,
	})

	return
}
