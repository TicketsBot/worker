package messagequeue

import (
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/config"
	"github.com/rxdn/gdl/cache"
)

func buildContext(ticket database.Ticket, cache *cache.PgCache) (ctx *worker.Context, err error) {
	ctx = &worker.Context{
		Cache:       cache,
		RateLimiter: nil, // Use http-proxy ratelimiting functionality
	}

	whitelabelBotId, isWhitelabel, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(ticket.GuildId)
	if err != nil {
		return ctx, err
	}

	ctx.IsWhitelabel = isWhitelabel

	if isWhitelabel {
		res, err := dbclient.Client.Whitelabel.GetByBotId(whitelabelBotId)
		if err != nil {
			return ctx, err
		}

		ctx.Token = res.Token
		ctx.BotId = whitelabelBotId
	} else {
		ctx.Token = config.Conf.Discord.Token

		ctx.BotId = config.Conf.Discord.PublicBotId
	}

	return ctx, err
}
