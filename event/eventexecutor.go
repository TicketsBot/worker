package event

import (
	"encoding/json"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/listeners"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/sirupsen/logrus"
	"reflect"
)

func execute(ctx *worker.Context, eventType events.EventType, data json.RawMessage) {
	dataType := events.EventTypes[eventType]
	if dataType == nil {
		return
	}

	event := reflect.New(dataType)
	if err := json.Unmarshal(data, event.Interface()); err != nil {
		logrus.Warnf("error whilst decoding event data: %s", err.Error())
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
				event,
			})
		}
	}
}

