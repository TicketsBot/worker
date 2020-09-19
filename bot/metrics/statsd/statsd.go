package statsd

import (
	stats "gopkg.in/alexcesaro/statsd.v2"
	"os"
	"sync"
	"time"
)

type StatsdClient struct {
	client *stats.Client
	buffer map[Key]int
	bufferLock sync.Mutex
}

var Client StatsdClient

func NewClient() (StatsdClient, error) {
	client, err := stats.New(stats.Address(os.Getenv("WORKER_STATSD_ADDR")), stats.Prefix(os.Getenv("WORKER_STATSD_PREFIX"))); if err != nil {
		return StatsdClient{}, err
	}

	return StatsdClient{
		client: client,
		buffer: make(map[Key]int),
	}, nil
}

func (c *StatsdClient) StartDaemon() {
	for {
		time.Sleep(time.Second * 15)

		c.bufferLock.Lock()
		for key, count := range c.buffer {
			c.client.Count(key.String(), count)
		}
		c.buffer = make(map[Key]int)
		c.bufferLock.Unlock()
	}
}

func IsClientNull() bool {
	return Client.client == nil
}

func (c *StatsdClient) IncrementKey(key Key) {
	if IsClientNull() {
		return
	}

	c.bufferLock.Lock()
	defer c.bufferLock.Unlock()

	var val int
	if current, ok := c.buffer[key]; ok {
		val = current
	} else {
		val = 0
	}

	val++
	c.buffer[key] = val

	c.client.Increment(key.String())
}
