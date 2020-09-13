package event

import (
	"encoding/json"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/listeners"
	"github.com/rxdn/gdl/gateway/payloads"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/sirupsen/logrus"
	"reflect"
)

func execute(ctx *worker.Context, event json.RawMessage) {
	var payload payloads.Payload
	if err := json.Unmarshal(event, &payload); err != nil {
		logrus.Warnf("error whilst decoding event data: %s", err.Error())
		return
	}

	dataType := events.EventTypes[events.EventType(payload.EventName)]
	if dataType == nil {
		return
	}

	data := reflect.New(dataType)
	if err := json.Unmarshal(payload.Data, data.Interface()); err != nil {
		logrus.Warnf("error whilst decoding event data: %s", err.Error())
		return
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

