package config

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	SentryDsn string `env:"WORKER_SENTRY_DSN"`
	DebugMode string `env:"WORKER_DEBUG"`

	Discord struct {
		Token       string `env:"WORKER_PUBLIC_TOKEN"`
		PublicBotId uint64 `env:"WORKER_PUBLIC_ID"`
		ProxyUrl    string `env:"DISCORD_PROXY_URL"`
	}

	Bot struct {
		HttpAddress         string   `env:"HTTP_ADDR"`
		SupportServerInvite string   `env:"SUPPORT_SERVER_INVITE"`
		Admins              []uint64 `env:"WORKER_BOT_ADMINS"`
		Helpers             []uint64 `env:"WORKER_BOT_HELPERS"`
	}

	PremiumProxy struct {
		Url string `env:"WORKER_PROXY_URL"`
		Key string `env:"WORKER_PROXY_KEY"`
	}

	Archiver struct {
		Url    string `env:"WORKER_ARCHIVER_URL"`
		AesKey string `env:"WORKER_ARCHIVER_AES_KEY"`
	}

	WebProxy struct {
		Url             string `env:"WEB_PROXY_URL"`
		AuthHeaderName  string `env:"WEB_PROXY_AUTH_HEADER_NAME"`
		AuthHeaderValue string `env:"WEB_PROXY_AUTH_HEADER_VALUE"`
	}

	Integrations struct {
		BloxlinkApiKey string `env:"BLOXLINK_API_KEY"`
	}

	Database struct {
		Host     string `env:"HOST"`
		Database string `env:"NAME"`
		Username string `env:"USER"`
		Password string `env:"PASSWORD"`
		Threads  int    `env:"THREADS"`
	} `envPrefix:"DATABASE_"`

	Cache struct {
		Host     string `env:"HOST"`
		Database string `env:"NAME"`
		Username string `env:"USER"`
		Password string `env:"PASSWORD"`
		Threads  int    `env:"THREADS"`
	} `envPrefix:"CACHE_"`

	Redis struct {
		Address  string `env:"WORKER_REDIS_ADDR"`
		Password string `env:"WORKER_REDIS_PASSWD"`
		Threads  int    `env:"WORKER_REDIS_THREADS"`
	}

	Statsd struct {
		Address string `env:"WORKER_STATSD_ADDR"`
		Prefix  string `env:"WORKER_STATSD_PREFIX"`
	}
}

var Conf Config

func Parse() {
	if err := env.Parse(&Conf); err != nil {
		panic(err)
	}
}
