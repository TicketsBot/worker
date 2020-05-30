package dbclient

import (
	"context"
	"github.com/TicketsBot/database"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var Client *database.Database

func Connect() {
	maxConns, err := strconv.Atoi(os.Getenv("WORKER_PG_THREADS"))
	if err != nil {
		panic(err)
	}

	config := &pgxpool.Config{
		ConnConfig: &pgx.ConnConfig{
			Config: pgconn.Config{
				Host:     os.Getenv("WORKER_PG_HOST"),
				Port:     5432,
				Database: os.Getenv("WORKER_PG_DATABASE"),
				User:     os.Getenv("WORKER_PG_USER"),
				Password: os.Getenv("WORKER_PG_PASSWD"),
			},
		},
		MaxConns: int32(maxConns),
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
