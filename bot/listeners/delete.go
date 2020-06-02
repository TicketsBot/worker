package listeners

import (
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnChannelDelete(worker *worker.Context, e *events.ChannelDelete, extra eventforwarding.Extra) {
	if err := dbclient.Client.Tickets.CloseByChannel(e.Id); err != nil {
		sentry.Error(err)
	}
}

