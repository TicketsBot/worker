package listeners

import (
	"context"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"time"
)

// Remove user permissions when they leave
func OnMemberUpdate(worker *worker.Context, e events.GuildMemberUpdate) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3) // TODO: Propagate context
	defer cancel()

	span := sentry.StartSpan(ctx, "OnMemberUpdate")
	defer span.Finish()

	if err := utils.ToRetriever(worker).Cache().DeleteCachedPermissionLevel(ctx, e.GuildId, e.User.Id); err != nil {
		sentry.Error(err)
	}
}
