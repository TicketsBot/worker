package listeners

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnReactionRemove(worker *worker.Context, e *events.MessageReactionRemoveAll) {
	panel, err := dbclient.Client.Panel.Get(e.MessageId)
	if err != nil {
		sentry.Error(err)
		return
	}

	if err := worker.CreateReaction(e.ChannelId, e.MessageId, panel.ReactionEmote); err != nil {
		sentry.Error(err)
	}
}
