package cache

import (
	"context"
	"github.com/jackc/pgconn"
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
	maxConns, err := strconv.Atoi(os.Getenv("CACHE_THREADS"))
	if err != nil {
		panic(err)
	}

	config := &pgxpool.Config{
		ConnConfig: &pgx.ConnConfig{
			Config: pgconn.Config{
				Host:     os.Getenv("CACHE_HOST"),
				Port:     5432,
				Database: os.Getenv("CACHE_NAME"),
				User:     os.Getenv("CACHE_USER"),
				Password: os.Getenv("CACHE_PASSWORD"),
			},
		},
		MaxConns: int32(maxConns),
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

