package worker

import (
	"context"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest/ratelimit"
)

type Context struct {
	context.Context
	Token        string
	BotId        uint64
	IsWhitelabel bool
	ShardId      int
	Cache        *cache.PgCache
	RateLimiter  *ratelimit.Ratelimiter
}

func (ctx *Context) Self() (user.User, error) {
	return ctx.GetUser(ctx.BotId)
}
