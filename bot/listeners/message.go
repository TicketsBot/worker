package listeners

import (
	"github.com/TicketsBot/common/chatrelay"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
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
	statsd.Client.IncrementKey(statsd.KeyMessages)

	// ignore DMs
	if e.GuildId == 0 {
		return
	}

	// Verify that this is a ticket
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(e.ChannelId, e.GuildId)
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
		// set participants, for logging
		if err := dbclient.Client.Participants.Set(e.GuildId, ticket.Id, e.Author.Id); err != nil {
			sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		}

		isStaff, err := isStaff(e, ticket)
		if err != nil {
			sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		} else {
			// set ticket last message, for autoclose
			if err := updateLastMessage(e, ticket, isStaff); err != nil {
				sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
			}

			// first response time
			// first, get if the user is staff
			e.Member.User = e.Author
			if err != nil {
				sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
			} else if isStaff { // check the user is staff
				// We don't have to check for previous responses due to ON CONFLICT DO NOTHING
				if err := dbclient.Client.FirstResponseTime.Set(e.GuildId, e.Author.Id, ticket.Id, time.Now().Sub(ticket.OpenTime)); err != nil {
					sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
				}
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

func updateLastMessage(msg *events.MessageCreate, ticket database.Ticket, isStaff bool) error {
	// If last message was sent by staff, don't reset the timer
	lastMessage, err := dbclient.Client.TicketLastMessage.Get(ticket.GuildId, ticket.Id)
	if err != nil {
		return err
	}

	// No last message, or last message was before we started storing user IDs
	if lastMessage.UserId == nil {
		return dbclient.Client.TicketLastMessage.Set(ticket.GuildId, ticket.Id, msg.Id, msg.Author.Id, false)
	}

	// If the last message was sent by the ticket opener, we can skip the rest of the logic, and update straight away
	if ticket.UserId == msg.Author.Id {
		return dbclient.Client.TicketLastMessage.Set(ticket.GuildId, ticket.Id, msg.Id, msg.Author.Id, false)
	}

	// If the last message *and* this message were sent by staff members, then do not reset the timer
	if lastMessage.UserIsStaff && isStaff {
		return nil
	}

	return dbclient.Client.TicketLastMessage.Set(ticket.GuildId, ticket.Id, msg.Id, msg.Author.Id, isStaff)
}

// This method should not be used for anything requiring elevated privileges
func isStaff(msg *events.MessageCreate, ticket database.Ticket) (bool, error) {
	// If the user is the ticket opener, they are not staff
	if msg.Author.Id == ticket.UserId {
		return false, nil
	}

	members, err := dbclient.Client.TicketMembers.Get(ticket.GuildId, ticket.Id)
	if err != nil {
		return false, err
	}

	if utils.Contains(members, msg.Author.Id) {
		return false, nil
	}

	return true, nil
}
