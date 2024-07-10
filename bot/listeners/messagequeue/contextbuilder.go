package messagequeue

import (
	"context"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/config"
	"github.com/rxdn/gdl/cache"
)

func buildContext(ctx context.Context, ticket database.Ticket, cache *cache.PgCache) (*worker.Context, error) {
	worker := &worker.Context{
		Cache:       cache,
		RateLimiter: nil, // Use http-proxy ratelimiting functionality
	}

	whitelabelBotId, isWhitelabel, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(ctx, ticket.GuildId)
	if err != nil {
		return nil, err
	}

	worker.IsWhitelabel = isWhitelabel

	if isWhitelabel {
		res, err := dbclient.Client.Whitelabel.GetByBotId(ctx, whitelabelBotId)
		if err != nil {
			return nil, err
		}

		worker.Token = res.Token
		worker.BotId = whitelabelBotId
	} else {
		worker.Token = config.Conf.Discord.Token
		worker.BotId = config.Conf.Discord.PublicBotId
	}

	return worker, err
}
