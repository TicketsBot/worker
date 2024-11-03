package listeners

import (
	"context"
	"errors"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/config"
	"github.com/rxdn/gdl/cache"
	"time"
)

type BaseListener struct {
	cache *cache.PgCache
}

const Timeout = time.Second * 15

func NewBaseListener(cache *cache.PgCache) *BaseListener {
	return &BaseListener{
		cache: cache,
	}
}

func (b *BaseListener) BuildContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), Timeout)
}

func (b *BaseListener) ContextForGuild(ctx context.Context, guildId uint64) (*worker.Context, error) {
	botId, isWhitelabel, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(ctx, guildId)
	if err != nil {
		return nil, err
	}

	if isWhitelabel {
		// TODO: Merge lookup into one query
		bot, err := dbclient.Client.Whitelabel.GetByBotId(ctx, botId)
		if err != nil {
			return nil, err
		}

		if bot.BotId == 0 {
			return nil, errors.New("bot not found")
		}

		return &worker.Context{
			Token:        bot.Token,
			BotId:        bot.BotId,
			IsWhitelabel: true,
			ShardId:      0,
			Cache:        b.cache,
			RateLimiter:  nil,
		}, nil
	} else {
		return &worker.Context{
			Token:        config.Conf.Discord.Token,
			BotId:        config.Conf.Discord.PublicBotId,
			IsWhitelabel: false,
			ShardId:      0,
			Cache:        b.cache,
			RateLimiter:  nil,
		}, nil
	}
}
