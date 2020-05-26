package worker

import (
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/rest/ratelimit"
)

type Context struct {
	Token       string
	BotId       uint64
	Cache       *cache.PgCache
	RateLimiter *ratelimit.Ratelimiter
}
