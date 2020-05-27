package main

import (
	"github.com/TicketsBot/archiverclient"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/event"
	"os"
)

func main() {
	utils.ParseBotAdmins()
	utils.ParseBotHelpers()

	redis.Connect()
	dbclient.Connect()

	cache, err := cache.Connect()
	if err != nil {
		panic(err)
	}

	utils.PremiumClient = premium.NewPremiumLookupClient(premium.NewPatreonClient(os.Getenv("WORKER_PROXY_URL"), os.Getenv("WORKER_PROXY_KEY")), redis.Client, &cache, dbclient.Client)
	utils.ArchiverClient = archiverclient.NewArchiverClient(os.Getenv("WORKER_ARCHIVER_URL"))

	statsd.Client, _ = statsd.NewClient()

	event.Listen(redis.Client, &cache)
}
