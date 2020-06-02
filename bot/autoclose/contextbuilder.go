package autoclose

import (
	"fmt"
	"github.com/TicketsBot/common/autoclose"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/rest/ratelimit"
	"os"
	"strconv"
)

func buildContext(ticket autoclose.Ticket, cache *cache.PgCache) (ctx *worker.Context, err error) {
	ctx = &worker.Context{
		Cache: cache,
	}

	whitelabelBotId, isWhitelabel, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(ticket.GuildId)
	if err != nil {
		return
	}

	ctx.IsWhitelabel = isWhitelabel

	var keyPrefix string

	if isWhitelabel {
		res, err := dbclient.Client.Whitelabel.GetByBotId(whitelabelBotId)
		if err != nil {
			return
		}

		ctx.Token = res.Token
		ctx.BotId = whitelabelBotId
		keyPrefix = fmt.Sprintf("ratelimiter:%d", whitelabelBotId)
	} else {
		ctx.Token = os.Getenv("WORKER_PUBLIC_TOKEN")
		keyPrefix = "ratelimiter:public"

		ctx.BotId, err = strconv.ParseUint(os.Getenv("WORKER_PUBLIC_ID"), 10, 64)
		if err != nil {
			return
		}
	}

	// TODO: Large sharding buckets
	ctx.RateLimiter = ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis.Client, keyPrefix), 1)

	return ctx, err
}
