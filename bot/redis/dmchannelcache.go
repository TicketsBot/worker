package redis

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/common/utils"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

var ErrNotCached = errors.New("channel not cached")

// Returns nil if we cannot create a channel
// Returns ErrNotCached if not cached
func GetDMChannel(userId, botId uint64) (*uint64, error) {
	key := fmt.Sprintf("dmchannel:%d:%d", botId, userId)

	res, err := Client.Get(utils.DefaultContext(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotCached
		}

		return nil, err
	}

	if res == "null" {
		return nil, nil
	}

	parsed, err := strconv.ParseUint(res, 10, 64)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func StoreNullDMChannel(userId, botId uint64) error {
	key := fmt.Sprintf("dmchannel:%d:%d", botId, userId)
	return Client.Set(utils.DefaultContext(), key, "null", time.Hour * 6).Err()
}

func StoreDMChannel(userId, channelId, botId uint64) error {
	key := fmt.Sprintf("dmchannel:%d:%d", botId, userId)
	return Client.Set(utils.DefaultContext(), key, strconv.FormatUint(channelId, 10), 0).Err()
}
