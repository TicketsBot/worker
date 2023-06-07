package dbclient

import (
	"context"
	"fmt"
	"github.com/TicketsBot/analytics-client"
	"github.com/TicketsBot/worker/config"
	"time"
)

var Analytics *analytics.Client

func ConnectAnalytics() {
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
		fmt.Printf("Clickhouse didn't respond to ping: %v\n", err)
		return
	}
}
