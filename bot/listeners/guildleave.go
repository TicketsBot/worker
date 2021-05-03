package listeners

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

/*
 * Sent when a guild becomes unavailable during a guild outage, or when the user leaves or is removed from a guild.
 * The inner payload is an unavailable guild object.
 * If the unavailable field is not set, the user was removed from the guild.
 */
func OnGuildLeave(worker *worker.Context, e *events.GuildDelete) {
	if e.Unavailable == nil {
		go statsd.Client.IncrementKey(statsd.KeyLeaves)

		if worker.IsWhitelabel {
			if err := dbclient.Client.WhitelabelGuilds.Delete(worker.BotId, e.Guild.Id); err != nil {
				sentry.Error(err)
			}
		}

		// Exclude from autoclose
		if err := dbclient.Client.AutoCloseExclude.ExcludeAll(e.Guild.Id); err != nil {
			sentry.Error(err)
		}

		if err := dbclient.Client.GuildLeaveTime.Set(e.Guild.Id); err != nil {
			sentry.Error(err)
		}
	}
}
