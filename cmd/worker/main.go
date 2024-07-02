package main

import (
	"fmt"
	"github.com/TicketsBot/archiverclient"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/integrations"
	"github.com/TicketsBot/worker/bot/listeners/messagequeue"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"github.com/TicketsBot/worker/event"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/rest/request"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	go func() {
		fmt.Println(http.ListenAndServe(":6060", nil))
	}()

	config.Parse()

	if len(config.Conf.DebugMode) == 0 {
		fmt.Println("Connecting to Sentry...")
		if err := sentry.Initialise(sentry.Options{
			Dsn:              config.Conf.Sentry.Dsn,
			Debug:            config.Conf.DebugMode != "",
			SampleRate:       config.Conf.Sentry.SampleRate,
			EnableTracing:    config.Conf.Sentry.UseTracing,
			TracesSampleRate: config.Conf.Sentry.TracingSampleRate,
		}); err != nil {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("Connected to Sentry, connect to Redis...")
	if err := redis.Connect(); err != nil {
		panic(err)
	}

	fmt.Println("Connected to Redis, connect to DB...")
	dbclient.Connect()

	i18n.Init()

	fmt.Println("Connected to DB, connect to cache...")
	pgCache, err := cache.Connect()
	if err != nil {
		panic(err)
	}

	cache.Client = &pgCache

	fmt.Println("Connected to cache, connect to clickhouse...")
	dbclient.ConnectAnalytics()

	// Configure HTTP proxy
	fmt.Println("Configuring proxy...")
	if config.Conf.Discord.ProxyUrl != "" {
		request.Client.Timeout = config.Conf.Discord.RequestTimeout
		request.RegisterPreRequestHook(utils.ProxyHook)
	}

	fmt.Println("Retrieved command list, initialising microservice clients...")
	if config.Conf.DebugMode == "" {
		utils.PremiumClient = premium.NewPremiumLookupClient(premium.NewPatreonClient(config.Conf.PremiumProxy.Url, config.Conf.PremiumProxy.Key), redis.Client, &pgCache, dbclient.Client)
	} else {
		c := premium.NewMockLookupClient(premium.Whitelabel, premium.SourcePatreon)
		utils.PremiumClient = &c

		request.Client.Timeout = time.Second * 10
	}

	utils.ArchiverClient = archiverclient.NewArchiverClient(config.Conf.Archiver.Url, []byte(config.Conf.Archiver.AesKey))

	prometheus.StartServer(config.Conf.Prometheus.Address)

	statsd.Client, err = statsd.NewClient(config.Conf.Statsd.Address, config.Conf.Statsd.Prefix)
	if err != nil {
		sentry.Error(err)
	} else {
		request.RegisterPreRequestHook(statsd.RestHook)
		go statsd.Client.StartDaemon()
	}

	request.RegisterPreRequestHook(prometheus.PreRequestHook)
	request.RegisterPostRequestHook(prometheus.PostRequestHook)

	integrations.InitIntegrations()

	go messagequeue.ListenTicketClose()
	go messagequeue.ListenAutoClose()
	go messagequeue.ListenCloseRequestTimer()

	fmt.Println("Listening for events...")
	event.HttpListen(redis.Client, &pgCache)
}
