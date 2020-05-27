package statsd

import (
	"fmt"
	stats "gopkg.in/alexcesaro/statsd.v2"
	"os"
)

type StatsdClient struct {
	*stats.Client
}

var Client StatsdClient

func NewClient() (StatsdClient, error) {
	addr := fmt.Sprintf("%s:%d", os.Getenv("WORKER_STATSD_HOST"), os.Getenv("WORKER_STATSD_PORT"))
	client, err := stats.New(stats.Address(addr), stats.Prefix(os.Getenv("WORKER_STATSD_PREFIX"))); if err != nil {
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
