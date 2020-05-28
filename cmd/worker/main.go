package main

import (
	"fmt"
	"github.com/TicketsBot/archiverclient"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/listeners/messagequeue"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/event"
	"os"
)

func main() {
	utils.ParseBotAdmins()
	utils.ParseBotHelpers()

	fmt.Println("Connect to redis...")
	if err := redis.Connect(); err != nil {
		panic(err)
	}

	fmt.Println("Connected to Redis, connect to DB...")
	dbclient.Connect()

	fmt.Println("Connected to DB, connect to cache...")
	pgCache, err := cache.Connect()
	if err != nil {
		panic(err)
	}

	cache.Client = &pgCache

	fmt.Println("Connected to cache, initialising microservice clients...")
	utils.PremiumClient = premium.NewPremiumLookupClient(premium.NewPatreonClient(os.Getenv("WORKER_PROXY_URL"), os.Getenv("WORKER_PROXY_KEY")), redis.Client, &pgCache, dbclient.Client)
	utils.ArchiverClient = archiverclient.NewArchiverClient(os.Getenv("WORKER_ARCHIVER_URL"))

	statsd.Client, _ = statsd.NewClient()

	go messagequeue.ListenTicketClose()

	fmt.Println("Listening for events...")
	event.Listen(redis.Client, &pgCache)
}
