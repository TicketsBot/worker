package dbclient

import (
	"context"
	"github.com/TicketsBot/analytics-client"
	"github.com/TicketsBot/worker/config"
	"go.uber.org/zap"
	"time"
)

var Analytics *analytics.Client

func ConnectAnalytics(logger *zap.Logger) {
	logger.Info("Connecting to Clickhouse",
		zap.String("address", config.Conf.Clickhouse.Address),
		zap.String("database", config.Conf.Clickhouse.Database),
		zap.String("username", config.Conf.Clickhouse.Username),
		zap.Int("threads", config.Conf.Clickhouse.Threads),
	)

	Analytics = analytics.Connect(
		config.Conf.Clickhouse.Address,
		config.Conf.Clickhouse.Threads,
		config.Conf.Clickhouse.Database,
		config.Conf.Clickhouse.Username,
		config.Conf.Clickhouse.Password,
		time.Second*10,
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := Analytics.Ping(ctx); err != nil {
		logger.Error("Clickhouse didn't response to ping", zap.Error(err))
		return
	}
}
