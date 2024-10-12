package redis

import (
	"context"
	"fmt"
	"time"
)

const channelRefetchBackoff = time.Minute * 5

func TakeChannelRefetchToken(ctx context.Context, guildId uint64) (bool, error) {
	key := fmt.Sprintf("channelrefetch:%d", guildId)
	return Client.SetNX(ctx, key, 1, channelRefetchBackoff).Result()
}
