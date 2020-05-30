package dbclient

import (
	"context"
	"github.com/TicketsBot/database"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"os"
)

var Client *database.Database

func Connect() {
	config, err := pgxpool.ParseConfig(os.Getenv("WORKER_PG_URI")); if err != nil {
		panic(err)
	}

	// TODO: Sentry
	config.ConnConfig.LogLevel = pgx.LogLevelWarn
	config.ConnConfig.Logger = logrusadapter.NewLogger(logrus.New())

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}

	Client = database.NewDatabase(pool)
	//Client.CreateTables(pool)
}

