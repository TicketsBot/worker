package dbclient

import (
	"context"
	"fmt"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/config"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

var Client *database.Database

func Connect(logger *zap.Logger) {
	cfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s/%s?pool_max_conns=%d",
		config.Conf.Database.Username,
		config.Conf.Database.Password,
		config.Conf.Database.Host,
		config.Conf.Database.Database,
		config.Conf.Database.Threads,
	))

	if err != nil {
		logger.Fatal("Failed to parse database config", zap.Error(err))
		return
	}

	// TODO: Sentry
	cfg.ConnConfig.LogLevel = pgx.LogLevelWarn
	cfg.ConnConfig.Logger = NewLogAdapter(logger)

	pool, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
		return
	}

	Client = database.NewDatabase(pool)
}
