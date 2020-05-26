package event

import (
	"fmt"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/worker"
	"github.com/go-redis/redis"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/rest/ratelimit"
)

func Listen(redis *redis.Client) {
	ch := make(chan eventforwarding.Event)
	go eventforwarding.Listen(redis, ch)

	for event := range ch {
		ctx := &worker.Context{
			Token:       event.BotToken,
			BotId:       event.BotId,
			Cache:       nil,
			RateLimiter: ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis, fmt.Sprintf("tickets:%d", event.BotId)), 1),
		}

		execute(ctx, events.EventType(event.EventType), event.Data)
	}
}
