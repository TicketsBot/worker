package statsd

import (
	stats "gopkg.in/alexcesaro/statsd.v2"
	"os"
)

type StatsdClient struct {
	*stats.Client
}

var Client StatsdClient

func NewClient() (StatsdClient, error) {
	client, err := stats.New(stats.Address(os.Getenv("WORKER_STATSD_ADDR")), stats.Prefix(os.Getenv("WORKER_STATSD_PREFIX"))); if err != nil {
		return StatsdClient{}, err
	}

	return StatsdClient{
		client,
	}, nil
}

func IsClientNull() bool {
	return Client.Client == nil
}

func IncrementKey(key Key) {
	if IsClientNull() {
		return
	}

	Client.Increment(key.String())
}
