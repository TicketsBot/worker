package main

import (
	"fmt"
	"github.com/TicketsBot/archiverclient"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/autoclose"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/listeners/messagequeue"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/event"
	"github.com/rxdn/gdl/rest/request"
	"os"
)

func main() {
	utils.ParseBotAdmins()
	utils.ParseBotHelpers()

	fmt.Println("Connecting to Sentry...")
	if err := sentry.Initialise(sentry.Options{
		Dsn:     os.Getenv("WORKER_SENTRY_DSN"),
		Project: "tickets-bot",
	}); err != nil {
		fmt.Println(err.Error())
	}

	// Configure HTTP proxy
	fmt.Println("Configuring proxy...")
	if os.Getenv("DISCORD_PROXY_URL") != "" {
		request.RegisterHook(utils.ProxyHook)
	}

	fmt.Println("Connected to Sentry, connect to Redis...")
	if err := redis.Connect(); err != nil {
		panic(err)
	}

	fmt.Println("Connected to Redis, connect to DB...")
	dbclient.Connect()

	if err := i18n.LoadMessages(dbclient.Client); err != nil {
		panic(err)
	}

	fmt.Println("Connected to DB, connect to cache...")
	pgCache, err := cache.Connect()
	if err != nil {
		panic(err)
	}

	cache.Client = &pgCache

	fmt.Println("Connected to cache, initialising microservice clients...")
	utils.PremiumClient = premium.NewPremiumLookupClient(premium.NewPatreonClient(os.Getenv("WORKER_PROXY_URL"), os.Getenv("WORKER_PROXY_KEY")), redis.Client, &pgCache, dbclient.Client)
	utils.ArchiverClient = archiverclient.NewArchiverClient(os.Getenv("WORKER_ARCHIVER_URL"), []byte(os.Getenv("WORKER_ARCHIVER_AES_KEY")))

	statsd.Client, err = statsd.NewClient()
	if err != nil {
		sentry.Error(err)
	} else {
		request.RegisterHook(statsd.RestHook)
		go statsd.Client.StartDaemon()
	}

	go messagequeue.ListenTicketClose()
	go autoclose.ListenAutoClose(&pgCache)

	fmt.Println("Listening for events...")
	event.Listen(redis.Client, &pgCache)
}
