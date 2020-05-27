package worker

import (
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest/ratelimit"
)

type Context struct {
	Token       string
	BotId       uint64
	ShardId     int
	Cache       *cache.PgCache
	RateLimiter *ratelimit.Ratelimiter
}

func (c *Context) Self() (user.User, error) {
	return c.GetUser(c.BotId)
}
