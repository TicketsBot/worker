package statsd

import (
	stats "gopkg.in/alexcesaro/statsd.v2"
	"os"
	"time"
)

type StatsdClient struct {
	client           *stats.Client
	buffer           map[Key]int
	incrementChannel chan Key
}

var Client StatsdClient

func NewClient() (StatsdClient, error) {
	client, err := stats.New(stats.Address(os.Getenv("WORKER_STATSD_ADDR")), stats.Prefix(os.Getenv("WORKER_STATSD_PREFIX")))
	if err != nil {
		return StatsdClient{}, err
	}

	return StatsdClient{
		client:           client,
		buffer:           make(map[Key]int),
		incrementChannel: make(chan Key),
	}, nil
}

func (c *StatsdClient) StartDaemon() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			for key, count := range c.buffer {
				c.client.Count(key.String(), count)
			}

			c.buffer = make(map[Key]int)
		case key := <-c.incrementChannel:
			c.buffer[key]++
		}
	}
}

func IsClientNull() bool {
	return Client.client == nil
}

func (c *StatsdClient) IncrementKey(key Key) {
	if IsClientNull() {
		return
	}

	c.incrementChannel <- key
}
