package cache

import (
	"context"
	"fmt"
	"github.com/TicketsBot/worker/config"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rxdn/gdl/cache"
	"github.com/sirupsen/logrus"
)

var Client *cache.PgCache

func Connect() (client cache.PgCache, err error) {
	cfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s/%s?pool_max_conns=%d",
		config.Conf.Cache.Username,
		config.Conf.Cache.Password,
		config.Conf.Cache.Host,
		config.Conf.Cache.Database,
		config.Conf.Cache.Threads,
	))

	if err != nil {
		panic(err)
	}

	// TODO: Sentry
	cfg.ConnConfig.LogLevel = pgx.LogLevelWarn
	cfg.ConnConfig.Logger = logrusadapter.NewLogger(logrus.New())
	cfg.ConnConfig.PreferSimpleProtocol = true

	pool, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return
	}

	client = cache.NewPgCache(pool, cache.CacheOptions{
		Guilds:      true,
		Users:       true,
		Members:     true,
		Channels:    true,
		Roles:       true,
		Emojis:      false,
		VoiceStates: false,
	})

	return
}
