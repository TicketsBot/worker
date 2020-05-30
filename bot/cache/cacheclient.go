package cache

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rxdn/gdl/cache"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var Client *cache.PgCache

func Connect() (client cache.PgCache, err error) {
	threads, err := strconv.Atoi(os.Getenv("CACHE_THREADS"))
	if err != nil {
		panic(err)
	}

	config, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s/%s?pool_max_conns=%d",
		os.Getenv("CACHE_USER"),
		os.Getenv("CACHE_PASSWORD"),
		os.Getenv("CACHE_HOST"),
		os.Getenv("CACHE_NAME"),
		threads,
	))

	if err != nil {
		panic(err)
	}

	// TODO: Sentry
	config.ConnConfig.LogLevel = pgx.LogLevelWarn
	config.ConnConfig.Logger = logrusadapter.NewLogger(logrus.New())

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
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
