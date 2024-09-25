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
	"go.uber.org/zap"
	"time"
)

var Client *cache.PgCache

func Connect(logger *zap.Logger) (client cache.PgCache, err error) {
	uri := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?pool_max_conns=%d",
		config.Conf.Cache.Username,
		config.Conf.Cache.Password,
		config.Conf.Cache.Host,
		config.Conf.Cache.Database,
		config.Conf.Cache.Threads,
	)

	cfg, err := pgxpool.ParseConfig(uri)
	if err != nil {
		panic(err)
	}

	logger.Info(
		"Connecting to Postgres cache",
		zap.String("username", config.Conf.Cache.Username),
		zap.String("host", config.Conf.Cache.Host),
		zap.String("database", config.Conf.Cache.Database),
		zap.Int("threads", config.Conf.Cache.Threads),
	)

	// TODO: Sentry
	cfg.ConnConfig.LogLevel = pgx.LogLevelWarn
	cfg.ConnConfig.Logger = logrusadapter.NewLogger(logrus.New())
	cfg.ConnConfig.PreferSimpleProtocol = true
	cfg.ConnConfig.ConnectTimeout = time.Second * 15

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return
	}

	client = cache.NewPgCache(pool, cache.CacheOptions{
		Guilds:      true,
		Users:       true,
		Members:     true,
		Channels:    true,
		Roles:       false,
		Emojis:      false,
		VoiceStates: false,
	})

	return
}
