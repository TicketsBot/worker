package listeners

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnChannelDelete(worker *worker.Context, e *events.ChannelDelete) {
	// if this is an ticket channel, close it
	if err := dbclient.Client.Tickets.CloseByChannel(e.Id); err != nil {
		sentry.Error(err)
	}

	// if this is a channel category, delete it
	if err := dbclient.Client.ChannelCategory.DeleteByChannel(e.Id); err != nil {
		sentry.Error(err)
	}

	// if this is an archive channel, delete it
	if err := dbclient.Client.ArchiveChannel.DeleteByChannel(e.Id); err != nil {
		sentry.Error(err)
	}
}

