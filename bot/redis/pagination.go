package redis

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

const expiry = time.Minute

func SetPage(client *redis.Client, msgId uint64, page int) {
	key := fmt.Sprintf("pagination:%d", msgId)
	client.Set(key, page, expiry)
}

func GetPage(client *redis.Client, msgId uint64) (page int, success bool) {
	key := fmt.Sprintf("pagination:%d", msgId)

	data, err := client.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return
		} else {
			sentry.Error(err)
			return
		}
	}

	page, err = strconv.Atoi(data)
	if err != nil {
		sentry.Error(err)
		return
	}

	success = true
	return
}
