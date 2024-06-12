package event

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/listeners"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/rxdn/gdl/gateway/payloads"
)

func execute(c *worker.Context, event []byte) error {
	var payload payloads.Payload
	if err := json.Unmarshal(event, &payload); err != nil {
		return errors.New(fmt.Sprintf("error whilst decoding event data: %s (data: %s)", err.Error(), string(event)))
	}

	if err := listeners.HandleEvent(c, payload); err != nil {
		return err
	}

	// Goroutine because recording metrics is blocking
	statsd.Client.IncrementKey(statsd.KeyEvents)

	return nil
}
