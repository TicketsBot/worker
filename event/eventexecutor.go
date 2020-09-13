package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/listeners"
	"github.com/rxdn/gdl/gateway/payloads"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"reflect"
)

func execute(ctx *worker.Context, event json.RawMessage) error {
	var payload payloads.Payload
	if err := json.Unmarshal(event, &payload); err != nil {
		return errors.New(fmt.Sprintf("error whilst decoding event data: %s (data: %s)", err.Error(), string(event)))
	}

	dataType := events.EventTypes[events.EventType(payload.EventName)]
	if dataType == nil {
		return errors.New(fmt.Sprintf("Invalid event type: %s", payload.EventName))
	}

	data := reflect.New(dataType)
	if err := json.Unmarshal(payload.Data, data.Interface()); err != nil {
		return errors.New(fmt.Sprintf("error whilst decoding event data: %s (data: %s)", err.Error(), string(event)))
	}

	for _, listener := range listeners.Listeners {
		fn := reflect.TypeOf(listener)
		if fn.NumIn() != 2 {
			continue
		}

		ptr := fn.In(1)
		if ptr.Kind() != reflect.Ptr {
			continue
		}

		if ptr.Elem() == dataType {
			go reflect.ValueOf(listener).Call([]reflect.Value{
				reflect.ValueOf(ctx),
				data,
			})
		}
	}
}

