package listeners

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

// Remove user permissions when they leave
func OnMemberUpdate(worker *worker.Context, e events.GuildMemberUpdate) {
	span := sentry.StartSpan(worker.Context, "OnMemberUpdate")
	defer span.Finish()

	if err := utils.ToRetriever(worker).Cache().DeleteCachedPermissionLevel(e.GuildId, e.User.Id); err != nil {
		sentry.Error(err)
	}
}
