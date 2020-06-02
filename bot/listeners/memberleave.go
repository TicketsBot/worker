package listeners

import (
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

// Remove user permissions when they leave
func OnMemberLeave(worker *worker.Context, e *events.GuildMemberRemove, extra eventforwarding.Extra) {
	if err := dbclient.Client.Permissions.RemoveSupport(e.GuildId, e.User.Id); err != nil {
		sentry.Error(err)
	}
}
