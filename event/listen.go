package event

import (
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/worker"
	"github.com/go-redis/redis"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/rest/ratelimit"
	"github.com/sirupsen/logrus"
)

func Listen(redis *redis.Client, cache *cache.PgCache) {
	ch := eventforwarding.Listen(redis)

	for event := range ch {
		var keyPrefix string

		if event.IsWhitelabel {
			keyPrefix = fmt.Sprintf("ratelimiter:%d", event.BotId)
		} else {
			keyPrefix = "ratelimiter:public"
		}

		ctx := &worker.Context{
			Token:        event.BotToken,
			BotId:        event.BotId,
			IsWhitelabel: event.IsWhitelabel,
			ShardId:      event.ShardId,
			Cache:        cache,
			RateLimiter:  ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis, keyPrefix), 1),
		}

		if err := execute(ctx, event.Event); err != nil {
			marshalled, _ := json.Marshal(event)
			logrus.Warnf("error executing event: %e (payload: %s)", err, string(marshalled))
		}
	}
}
