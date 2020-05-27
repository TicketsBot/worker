package event

import (
	"encoding/json"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/listeners"
	modmaillisteners "github.com/TicketsBot/worker/bot/modmail/listeners"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/sirupsen/logrus"
	"reflect"
)

var allListeners = append(listeners.Listeners, modmaillisteners.Listeners...)

func execute(ctx *worker.Context, eventType events.EventType, data json.RawMessage, extra eventforwarding.Extra) {
	dataType := events.EventTypes[eventType]
	if dataType == nil {
		return
	}

	event := reflect.New(dataType)
	if err := json.Unmarshal(data, event.Interface()); err != nil {
		logrus.Warnf("error whilst decoding event data: %s", err.Error())
	}

	for _, listener := range allListeners {
		fn := reflect.TypeOf(listener)
		if fn.NumIn() != 3 {
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
				reflect.ValueOf(extra),
			})
		}
	}
}

