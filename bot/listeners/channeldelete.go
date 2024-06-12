package listeners

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnChannelDelete(worker *worker.Context, e events.ChannelDelete) {
	// If this is a ticket channel, close it
	if err := sentry.WithSpan1(worker.Context, "Close ticket by channel", func(span *sentry.Span) error {
		return dbclient.Client.Tickets.CloseByChannel(e.Id)
	}); err != nil {
		sentry.Error(err)
	}

	// if this is a channel category, delete it
	if err := sentry.WithSpan1(worker.Context, "Delete category by channel", func(span *sentry.Span) error {
		return dbclient.Client.ChannelCategory.DeleteByChannel(e.Id)
	}); err != nil {
		sentry.Error(err)
	}

	// if this is an archive channel, delete it
	if err := sentry.WithSpan1(worker.Context, "Delete archive channel by channel", func(span *sentry.Span) error {
		return dbclient.Client.ArchiveChannel.DeleteByChannel(e.Id)
	}); err != nil {
		sentry.Error(err)
	}
}
