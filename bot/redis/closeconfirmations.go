package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

const timeout = time.Second * 10

func SetCloseConfirmation(redis *redis.Client, messageId, userId uint64) error {
	key := fmt.Sprintf("closeconfirmation:%d:%d", messageId, userId)
	return redis.Set(key, 1, timeout).Err()
}

// returns if there was a pending close confirmation
func ConfirmClose(redis *redis.Client, messageId, userId uint64) bool {
	key := fmt.Sprintf("closeconfirmation:%d:%d", messageId, userId)
	res, err := redis.Del(key).Result()
	if err != nil {
		return false
	}

	return res > 0
}
