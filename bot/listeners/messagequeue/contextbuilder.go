package messagequeue

import (
	"fmt"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/config"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/rest/ratelimit"
)

func buildContext(ticket database.Ticket, cache *cache.PgCache) (ctx *worker.Context, err error) {
	ctx = &worker.Context{
		Cache: cache,
	}

	whitelabelBotId, isWhitelabel, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(ticket.GuildId)
	if err != nil {
		return ctx, err
	}

	ctx.IsWhitelabel = isWhitelabel

	var keyPrefix string

	if isWhitelabel {
		res, err := dbclient.Client.Whitelabel.GetByBotId(whitelabelBotId)
		if err != nil {
			return ctx, err
		}

		ctx.Token = res.Token
		ctx.BotId = whitelabelBotId
		keyPrefix = fmt.Sprintf("ratelimiter:%d", whitelabelBotId)
	} else {
		ctx.Token = config.Conf.Discord.Token
		keyPrefix = "ratelimiter:public"

		ctx.BotId = config.Conf.Discord.PublicBotId
	}

	// TODO: Large sharding buckets
	ctx.RateLimiter = ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis.Client, keyPrefix), 1)

	return ctx, err
}
