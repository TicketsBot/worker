package redis

import (
	"github.com/TicketsBot/worker/config"
	"github.com/go-redis/redis/v8"
)

var Client *redis.Client

func Connect() error {
	Client = redis.NewClient(&redis.Options{
		Network:      "tcp",
		Addr:         config.Conf.Redis.Address,
		Password:     config.Conf.Redis.Password,
		PoolSize:     config.Conf.Redis.Threads,
		MinIdleConns: config.Conf.Redis.Threads,
	})

	return nil
}
