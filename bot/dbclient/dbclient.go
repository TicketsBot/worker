package dbclient

import (
	"context"
	"fmt"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/config"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

var Client *database.Database
var Pool *pgxpool.Pool

func Connect() {
	fmt.Println(fmt.Sprintf(
		"postgres://%s:%s@%s/%s?pool_max_conns=%d",
		config.Conf.Database.Username,
		config.Conf.Database.Password,
		config.Conf.Database.Host,
		config.Conf.Database.Database,
		config.Conf.Database.Threads,
	))
	cfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s/%s?pool_max_conns=%d",
		config.Conf.Database.Username,
		config.Conf.Database.Password,
		config.Conf.Database.Host,
		config.Conf.Database.Database,
		config.Conf.Database.Threads,
	))

	if err != nil {
		panic(err)
	}

	// TODO: Sentry
	cfg.ConnConfig.LogLevel = pgx.LogLevelWarn
	cfg.ConnConfig.Logger = logrusadapter.NewLogger(logrus.New())

	Pool, err = pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	Client = database.NewDatabase(Pool)
}
