package statsd

import (
	"go.uber.org/atomic"
	stats "gopkg.in/alexcesaro/statsd.v2"
	"sync"
	"time"
)

type StatsdClient struct {
	client *stats.Client
	buffer map[Key]*atomic.Int32
	mu     *sync.Mutex
}

var Client StatsdClient

func NewClient(statsdAddress, statsdPrefix string) (StatsdClient, error) {
	client, err := stats.New(stats.Address(statsdAddress), stats.Prefix(statsdPrefix))
	if err != nil {
		return StatsdClient{}, err
	}

	buffer := make(map[Key]*atomic.Int32)
	for _, key := range AllKeys() {
		buffer[key] = atomic.NewInt32(0)
	}

	return StatsdClient{
		client: client,
		buffer: buffer,
		mu:     &sync.Mutex{},
	}, nil
}

func (c *StatsdClient) StartDaemon() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			for key, count := range c.buffer {
				c.client.Count(key.String(), count.Swap(0))
			}
		}
	}
}

func (c *StatsdClient) IncrementKey(key Key) {
	if c.buffer[key] == nil {
		return
	}

	c.buffer[key].Inc()
}
