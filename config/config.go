package config

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
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
		SecureProxyUrl string `env:"SECURE_PROXY_URL"`
	}

	Database struct {
		Host     string `env:"HOST"`
		Database string `env:"NAME"`
		Username string `env:"USER"`
		Password string `env:"PASSWORD"`
		Threads  int    `env:"THREADS"`
	} `envPrefix:"DATABASE_"`

	Clickhouse struct {
		Address  string `env:"ADDR"`
		Threads  int    `env:"THREADS"`
		Database string `env:"DATABASE"`
		Username string `env:"USERNAME"`
		Password string `env:"PASSWORD"`
	} `envPrefix:"CLICKHOUSE_"`

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

	Prometheus struct {
		Address string `env:"PROMETHEUS_SERVER_ADDR"`
	}

	Statsd struct {
		Address string `env:"WORKER_STATSD_ADDR"`
		Prefix  string `env:"WORKER_STATSD_PREFIX"`
	}

	Sentry struct {
		Dsn               string  `env:"DSN"`
		UseTracing        bool    `env:"TRACING_ENABLED"`
		TracingSampleRate float64 `env:"TRACING_SAMPLE_RATE"`
	} `envPrefix:"WORKER_SENTRY_"`
}

var Conf Config

func Parse() {
	if err := env.Parse(&Conf); err != nil {
		panic(err)
	}
}
