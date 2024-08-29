package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/google/uuid"
	"time"
)

type Config struct {
	DebugMode   string `env:"WORKER_DEBUG"`
	PremiumOnly bool   `env:"WORKER_PREMIUM_ONLY" envDefault:"false"`

	Discord struct {
		Token            string        `env:"WORKER_PUBLIC_TOKEN"`
		PublicBotId      uint64        `env:"WORKER_PUBLIC_ID"`
		ProxyUrl         string        `env:"DISCORD_PROXY_URL"`
		RequestTimeout   time.Duration `env:"DISCORD_REQUEST_TIMEOUT" envDefault:"15s"`
		CallbackTimeout  time.Duration `env:"DISCORD_CALLBACK_TIMEOUT" envDefault:"2000ms"`
		DeferHardTimeout time.Duration `env:"DISCORD_DEFER_HARD_TIMEOUT" envDefault:"2500ms"`
	}

	Bot struct {
		HttpAddress         string   `env:"HTTP_ADDR"`
		SupportServerInvite string   `env:"SUPPORT_SERVER_INVITE"`
		Admins              []uint64 `env:"WORKER_BOT_ADMINS"`
		Helpers             []uint64 `env:"WORKER_BOT_HELPERS"`
	}

	PremiumProxy struct {
		Url string `env:"URL"`
		Key string `env:"KEY"`
	} `envPrefix:"WORKER_PROXY_"`

	Archiver struct {
		Url    string `env:"URL"`
		AesKey string `env:"AES_KEY"`
	} `envPrefix:"WORKER_ARCHIVER_"`

	WebProxy struct {
		Url             string `env:"URL"`
		AuthHeaderName  string `env:"AUTH_HEADER_NAME"`
		AuthHeaderValue string `env:"AUTH_HEADER_VALUE"`
	} `envPrefix:"WEB_PROXY_"`

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
		Address  string `env:"ADDR"`
		Password string `env:"PASSWD"`
		Threads  int    `env:"THREADS"`
	} `envPrefix:"WORKER_REDIS_"`

	Prometheus struct {
		Address string `env:"PROMETHEUS_SERVER_ADDR"`
	}

	Statsd struct {
		Address string `env:"ADDR"`
		Prefix  string `env:"PREFIX"`
	} `envPrefix:"WORKER_STATSD_"`

	Sentry struct {
		Dsn               string  `env:"DSN"`
		SampleRate        float64 `env:"SAMPLE_RATE" envDefault:"1.0"`
		UseTracing        bool    `env:"TRACING_ENABLED"`
		TracingSampleRate float64 `env:"TRACING_SAMPLE_RATE"`
	} `envPrefix:"WORKER_SENTRY_"`

	CloudProfiler struct {
		Enabled   bool   `env:"ENABLED" envDefault:"false"`
		ProjectId string `env:"PROJECT_ID"`
	} `envPrefix:"WORKER_CLOUD_PROFILER_"`

	VoteSkuId uuid.UUID `env:"VOTE_SKU_ID"`
}

var Conf Config

func Parse() {
	if err := env.Parse(&Conf); err != nil {
		panic(err)
	}
}
