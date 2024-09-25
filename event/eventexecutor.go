package event

import (
	"context"
	"errors"
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/listeners"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/getsentry/sentry-go"
	"github.com/rxdn/gdl/gateway/payloads"
)

func execute(c *worker.Context, event []byte) error {
	var payload payloads.Payload
	if err := json.Unmarshal(event, &payload); err != nil {
		return errors.New(fmt.Sprintf("error whilst decoding event data: %s (data: %s)", err.Error(), string(event)))
	}

	span := sentry.StartTransaction(context.Background(), "Handle Event")
	span.SetTag("event", payload.EventName)
	defer span.Finish()

	prometheus.Events.WithLabelValues(payload.EventName).Inc()

	if err := listeners.HandleEvent(c, span, payload); err != nil {
		return err
	}

	return nil
}
