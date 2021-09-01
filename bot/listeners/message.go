package listeners

import (
	"github.com/TicketsBot/common/chatrelay"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"time"
)

// proxy messages to web UI + set last message id
func OnMessage(worker *worker.Context, e *events.MessageCreate) {
	go statsd.Client.IncrementKey(statsd.KeyMessages)

	// ignore DMs
	if e.GuildId == 0 {
		return
	}

	// Verify that this is a ticket
	ticket, err := dbclient.Client.Tickets.GetByChannel(e.ChannelId)
	if err != nil {
		sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		return
	}

	// ensure valid ticket channel
	if ticket.Id == 0 {
		return
	}

	// ignore our own messages
	if e.Author.Id != worker.BotId && !e.Author.Bot {
		// set ticket last message, for autoclose
		if err := dbclient.Client.TicketLastMessage.Set(e.GuildId, ticket.Id, e.Id); err != nil {
			sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		}

		// set participants, for logging
		if err := dbclient.Client.Participants.Set(e.GuildId, ticket.Id, e.Author.Id); err != nil {
			sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		}

		// first response time
		// first, get if the user is staff
		e.Member.User = e.Author
		permLevel, err := permission.GetPermissionLevel(utils.ToRetriever(worker), e.Member, e.GuildId)
		if err != nil {
			sentry.Error(err)
		} else if permLevel > permission.Everyone { // check the user is staff
			// We don't have to check for previous responses due to ON CONFLICT DO NOTHING
			if err := dbclient.Client.FirstResponseTime.Set(e.GuildId, e.Author.Id, ticket.Id, time.Now().Sub(ticket.OpenTime)); err != nil {
				sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
			}
		}
	}

	premiumTier, err := utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter)
	if err != nil {
		sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		return
	}

	// proxy msg to web UI
	if premiumTier > premium.None {
		data := chatrelay.MessageData{
			Ticket:  ticket,
			Message: e.Message,
		}

		if err := chatrelay.PublishMessage(redis.Client, data); err != nil {
			sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		}
	}
}
