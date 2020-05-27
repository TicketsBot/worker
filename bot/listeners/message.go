package listeners

import (
	"github.com/TicketsBot/common/chatrelay"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

// proxy messages to web UI
func OnMessage(worker *worker.Context, e *events.MessageCreate, extra eventforwarding.Extra) {
	go statsd.IncrementKey(statsd.MESSAGES)

	// ignore DMs
	if e.GuildId == 0 {
		return
	}

	if utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter) > premium.None {
		ticket, err := dbclient.Client.Tickets.GetByChannel(e.ChannelId)
		if err != nil {
			sentry.Error(err)
			return
		}

		// Verify that this is a ticket
		if ticket.UserId != 0 {
			data := chatrelay.MessageData{
				Ticket:  ticket,
				Message: e.Message,
			}

			if err := chatrelay.PublishMessage(redis.Client, data); err != nil {
				sentry.Error(err)
			}
		}
	}
}

