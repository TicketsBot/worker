package event

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/listeners"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/rxdn/gdl/gateway/payloads"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"reflect"
)

func execute(ctx *worker.Context, event []byte) error {
	var payload payloads.Payload
	if err := json.Unmarshal(event, &payload); err != nil {
		return errors.New(fmt.Sprintf("error whilst decoding event data: %s (data: %s)", err.Error(), string(event)))
	}

	dataType := events.EventTypes[events.EventType(payload.EventName)]
	if dataType == nil {
		return fmt.Errorf("Invalid event type: %s", payload.EventName)
	}

	data := reflect.New(dataType)
	if err := json.Unmarshal(payload.Data, data.Interface()); err != nil {
		return fmt.Errorf("error whilst decoding event data: %s (data: %s)", err.Error(), string(event))
	}

	listeners, ok := listeners.Listeners[events.EventType(payload.EventName)]
	if ok { // Verify we have listeners registered for this event type
		for _, listener := range listeners {
			go reflect.ValueOf(listener).Call([]reflect.Value{
				reflect.ValueOf(ctx),
				data,
			})
		}
	}

	// Goroutine because recording metrics is blocking
	go statsd.Client.IncrementKey(statsd.KeyEvents)

	return nil
}
