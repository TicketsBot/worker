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
	maxConns, err := strconv.Atoi(os.Getenv("DATABASE_THREADS"))
	if err != nil {
		panic(err)
	}

	config := &pgxpool.Config{
		ConnConfig: &pgx.ConnConfig{
			Config: pgconn.Config{
				Host:     os.Getenv("DATABASE_HOST"),
				Port:     5432,
				Database: os.Getenv("DATABASE_NAME"),
				User:     os.Getenv("DATABASE_USER"),
				Password: os.Getenv("DATABASE_PASSWORD"),
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
