package listeners

import (
	"context"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"time"
)

/*
 * Sent when a guild becomes unavailable during a guild outage, or when the user leaves or is removed from a guild.
 * The inner payload is an unavailable guild object.
 * If the unavailable field is not set, the user was removed from the guild.
 */
func OnGuildLeave(worker *worker.Context, e events.GuildDelete) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3) // TODO: Propagate context
	defer cancel()

	span := sentry.StartSpan(ctx, "OnGuildLeave")
	defer span.Finish()

	if e.Unavailable == nil {
		statsd.Client.IncrementKey(statsd.KeyLeaves)

		if worker.IsWhitelabel {
			if err := dbclient.Client.WhitelabelGuilds.Delete(ctx, worker.BotId, e.Guild.Id); err != nil {
				sentry.Error(err)
			}
		}

		// Exclude from autoclose
		if err := dbclient.Client.AutoCloseExclude.ExcludeAll(ctx, e.Guild.Id); err != nil {
			sentry.Error(err)
		}

		if err := dbclient.Client.GuildLeaveTime.Set(ctx, e.Guild.Id); err != nil {
			sentry.Error(err)
		}
	}
}
