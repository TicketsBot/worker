package event

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/listeners"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/getsentry/sentry-go"
	"github.com/rxdn/gdl/gateway/payloads"
)

func execute(c *worker.Context, event []byte) error {
	var payload payloads.Payload
	if err := json.Unmarshal(event, &payload); err != nil {
		return errors.New(fmt.Sprintf("error whilst decoding event data: %s (data: %s)", err.Error(), string(event)))
	}

	span := sentry.StartTransaction(c.Context, "Handle Event")
	span.SetTag("event", payload.EventName)
	defer span.Finish()

	// TODO: This might be bad
	c.Context = span.Context()

	if err := listeners.HandleEvent(c, span, payload); err != nil {
		return err
	}

	// Goroutine because recording metrics is blocking
	statsd.Client.IncrementKey(statsd.KeyEvents)

	return nil
}
