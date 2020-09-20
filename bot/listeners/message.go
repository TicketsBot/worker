package listeners

import (
	"github.com/TicketsBot/common/chatrelay"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

// proxy messages to web UI + set last message id
func OnMessage(worker *worker.Context, e *events.MessageCreate) {
	go statsd.Client.IncrementKey(statsd.MESSAGES)

	// ignore DMs
	if e.GuildId == 0 {
		return
	}

	// Verify that this is a ticket
	ticket, err := dbclient.Client.Tickets.GetByChannel(e.ChannelId)
	if err != nil {
		sentry.Error(err)
		return
	}

	if ticket.UserId != 0 {
		// ignore our own messages
		if e.Author.Id != worker.BotId {
			if err := dbclient.Client.TicketLastMessage.Set(e.GuildId, ticket.Id, e.Id); err != nil {
				sentry.Error(err)
			}

			if err := dbclient.Client.Participants.Set(e.GuildId, ticket.Id, e.Id); err != nil {
				sentry.Error(err)
			}
		}

		// proxy msg to web UI
		if utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter) > premium.None {
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

