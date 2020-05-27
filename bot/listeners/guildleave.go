package listeners

import (
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

/*
 * Sent when a guild becomes unavailable during a guild outage, or when the user leaves or is removed from a guild.
 * The inner payload is an unavailable guild object.
 * If the unavailable field is not set, the user was removed from the guild.
 */
func OnGuildLeave(worker *worker.Context, e *events.GuildDelete, extra eventforwarding.Extra) {
	if e.Unavailable == nil {
		go statsd.IncrementKey(statsd.LEAVES)

		if worker.IsWhitelabel {
			if err := dbclient.Client.WhitelabelGuilds.Delete(worker.BotId, e.Guild.Id); err != nil {
				sentry.Error(err)
			}
		}
	}
}
