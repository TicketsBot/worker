package main

import (
	"fmt"
	"github.com/TicketsBot/archiverclient"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/listeners/messagequeue"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/event"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/ratelimit"
	"github.com/rxdn/gdl/rest/request"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"
)

func main() {
	go func() {
		fmt.Println(http.ListenAndServe(":6060", nil))
	}()

	utils.ParseBotAdmins()
	utils.ParseBotHelpers()

	fmt.Println("Connecting to Sentry...")
	if err := sentry.Initialise(sentry.Options{
		Dsn:     os.Getenv("WORKER_SENTRY_DSN"),
		Project: "tickets-bot",
		Debug:   os.Getenv("WORKER_DEBUG") != "",
	}); err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Connected to Sentry, connect to Redis...")
	if err := redis.Connect(); err != nil {
		panic(err)
	}

	fmt.Println("Connected to Redis, connect to DB...")
	dbclient.Connect()

	i18n.LoadMessages()

	fmt.Println("Connected to DB, connect to cache...")
	pgCache, err := cache.Connect()
	if err != nil {
		panic(err)
	}

	cache.Client = &pgCache

	fmt.Println("Connected to cache, retrieving command list...")
	{
		token := os.Getenv("WORKER_PUBLIC_TOKEN")
		ratelimiter := ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis.Client, "ratelimiter:public"), 1)
		botId, err := strconv.ParseUint(os.Getenv("WORKER_PUBLIC_ID"), 10, 64)
		if err != nil {
			panic(err)
		}

		bot.GlobalCommands, err = rest.GetGlobalCommands(token, ratelimiter, botId)
		if err != nil {
			panic(err)
		}
	}

	// Configure HTTP proxy
	fmt.Println("Configuring proxy...")
	if os.Getenv("DISCORD_PROXY_URL") != "" {
		request.Client.Timeout = time.Second * 30
		request.RegisterHook(utils.ProxyHook)
	}

	fmt.Println("Retrieved command list, initialising microservice clients...")
	if os.Getenv("WORKER_DEBUG") == "" {
		utils.PremiumClient = premium.NewPremiumLookupClient(premium.NewPatreonClient(os.Getenv("WORKER_PROXY_URL"), os.Getenv("WORKER_PROXY_KEY")), redis.Client, &pgCache, dbclient.Client)
	} else {
		c := premium.NewMockLookupClient(premium.Whitelabel, premium.SourcePatreon)
		utils.PremiumClient = &c
	}

	utils.ArchiverClient = archiverclient.NewArchiverClient(os.Getenv("WORKER_ARCHIVER_URL"), []byte(os.Getenv("WORKER_ARCHIVER_AES_KEY")))

	statsd.Client, err = statsd.NewClient()
	if err != nil {
		sentry.Error(err)
	} else {
		request.RegisterHook(statsd.RestHook)
		go statsd.Client.StartDaemon()
	}

	go messagequeue.ListenTicketClose()
	go messagequeue.ListenAutoClose()

	fmt.Println("Listening for events...")
	event.HttpListen(redis.Client, &pgCache)
}
