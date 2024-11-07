package listeners

import (
	"context"
	"errors"
	"github.com/TicketsBot/common/chatrelay"
	"github.com/TicketsBot/common/model"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"strconv"
	"time"
)

// proxy messages to web UI + set last message id
func OnMessage(worker *worker.Context, e events.MessageCreate) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*7) // TODO: Propagate context
	defer cancel()

	span := sentry.StartTransaction(ctx, "OnMessage")
	defer span.Finish()

	if e.GuildId != 0 {
		span.SetTag("guild_id", strconv.FormatUint(e.GuildId, 10))
	}

	statsd.Client.IncrementKey(statsd.KeyMessages)

	// ignore DMs
	if e.GuildId == 0 {
		return
	}

	ticket, isTicket, err := getTicket(span.Context(), e.ChannelId)
	if err != nil {
		sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		return
	}

	// ensure valid ticket channel
	if !isTicket || ticket.Id == 0 {
		return
	}

	var isStaffCached *bool

	// ignore our own messages
	if e.Author.Id != worker.BotId && !e.Author.Bot {
		// set participants, for logging
		sentry.WithSpan0(span.Context(), "Add participant", func(span *sentry.Span) {
			if err := dbclient.Client.Participants.Set(ctx, e.GuildId, ticket.Id, e.Author.Id); err != nil {
				sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
			}
		})

		isStaffCached, err = sentry.WithSpan2(span.Context(), "Update ticket last activity", func(span *sentry.Span) (*bool, error) {
			v, err := isStaff(ctx, e, ticket)
			return &v, err
		})

		if err != nil {
			sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		} else {
			// set ticket last message, for autoclose
			// isStaffCached cannot be nil at this point
			if err := updateLastMessage(span.Context(), e, ticket, *isStaffCached); err != nil {
				sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
			}

			if *isStaffCached { // check the user is staff
				// We don't have to check for previous responses due to ON CONFLICT DO NOTHING
				sentry.WithSpan0(span.Context(), "Set first response time", func(span *sentry.Span) {
					if err := dbclient.Client.FirstResponseTime.Set(ctx, e.GuildId, e.Author.Id, ticket.Id, time.Now().Sub(ticket.OpenTime)); err != nil {
						sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
					}
				})
			}
		}
	}

	premiumTier, err := sentry.WithSpan2(span.Context(), "Get premium tier", func(span *sentry.Span) (premium.PremiumTier, error) {
		return utils.PremiumClient.GetTierByGuildId(ctx, e.GuildId, true, worker.Token, worker.RateLimiter)
	})
	if err != nil {
		sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		return
	}

	// proxy msg to web UI
	if premiumTier > premium.None {
		if err := sentry.WithSpan1(span.Context(), "Relay message to dashboard", func(span *sentry.Span) error {
			data := chatrelay.MessageData{
				Ticket:  ticket,
				Message: e.Message,
			}

			prometheus.ForwardedDashboardMessages.Inc()

			return chatrelay.PublishMessage(redis.Client, data)
		}); err != nil {
			sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
		}

		// Ignore the welcome message and ping message
		if e.Author.Id != worker.BotId {
			var userIsStaff bool
			if isStaffCached != nil {
				userIsStaff = *isStaffCached
			} else {
				tmp, err := isStaff(ctx, e, ticket)
				if err != nil {
					sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
					return
				}

				userIsStaff = tmp
			}

			var newStatus model.TicketStatus
			if userIsStaff {
				newStatus = model.TicketStatusPending
			} else {
				newStatus = model.TicketStatusOpen
			}

			if ticket.Status != newStatus {
				if err := dbclient.Client.Tickets.SetStatus(ctx, e.GuildId, ticket.Id, newStatus); err != nil {
					sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
				}

				if !ticket.IsThread {
					if err := sentry.WithSpan1(span.Context(), "Update status update queue", func(span *sentry.Span) error {
						return dbclient.Client.CategoryUpdateQueue.Add(ctx, e.GuildId, ticket.Id, newStatus)
					}); err != nil {
						sentry.ErrorWithContext(err, utils.MessageCreateErrorContext(e))
					}
				}
			}
		}
	}
}

func updateLastMessage(ctx context.Context, msg events.MessageCreate, ticket database.Ticket, isStaff bool) error {
	span := sentry.StartSpan(ctx, "Update last message")
	defer span.Finish()

	// If last message was sent by staff, don't reset the timer
	lastMessage, err := dbclient.Client.TicketLastMessage.Get(ctx, ticket.GuildId, ticket.Id)
	if err != nil {
		return err
	}

	// No last message, or last message was before we started storing user IDs
	if lastMessage.UserId == nil {
		return dbclient.Client.TicketLastMessage.Set(ctx, ticket.GuildId, ticket.Id, msg.Id, msg.Author.Id, false)
	}

	// If the last message was sent by the ticket opener, we can skip the rest of the logic, and update straight away
	if ticket.UserId == msg.Author.Id {
		return dbclient.Client.TicketLastMessage.Set(ctx, ticket.GuildId, ticket.Id, msg.Id, msg.Author.Id, false)
	}

	// If the last message *and* this message were sent by staff members, then do not reset the timer
	if lastMessage.UserId != nil && *lastMessage.UserIsStaff && isStaff {
		return nil
	}

	return dbclient.Client.TicketLastMessage.Set(ctx, ticket.GuildId, ticket.Id, msg.Id, msg.Author.Id, isStaff)
}

// This method should not be used for anything requiring elevated privileges
func isStaff(ctx context.Context, msg events.MessageCreate, ticket database.Ticket) (bool, error) {
	// If the user is the ticket opener, they are not staff
	if msg.Author.Id == ticket.UserId {
		return false, nil
	}

	members, err := dbclient.Client.TicketMembers.Get(ctx, ticket.GuildId, ticket.Id)
	if err != nil {
		return false, err
	}

	if utils.Contains(members, msg.Author.Id) {
		return false, nil
	}

	return true, nil
}

func getTicket(ctx context.Context, channelId uint64) (database.Ticket, bool, error) {
	isTicket, err := sentry.WithSpan2(ctx, "IsTicketChannel redis lookup", func(span *sentry.Span) (bool, error) {
		return redis.IsTicketChannel(ctx, channelId)
	})

	cacheHit := err == nil

	if err == nil && !isTicket {
		prometheus.LogOnMessageTicketLookup(false, cacheHit)
		return database.Ticket{}, false, nil
	} else if err != nil && !errors.Is(err, redis.ErrTicketStatusNotCached) {
		return database.Ticket{}, false, err
	}

	// Either cache miss or the ticket *does* exist, so we need to fetch the object from the database
	ticket, err := sentry.WithSpan2(ctx, "Get ticket by channel", func(span *sentry.Span) (database.Ticket, error) {
		ticket, ok, err := dbclient.Client.Tickets.GetByChannel(ctx, channelId)
		if err != nil {
			return database.Ticket{}, err
		}

		if !ok {
			return database.Ticket{}, nil
		}

		return ticket, nil
	})

	if err != nil {
		return database.Ticket{}, false, err
	}

	if err := redis.SetTicketChannelStatus(ctx, channelId, ticket.Id != 0); err != nil {
		return database.Ticket{}, false, err
	}

	if ticket.Id == 0 {
		prometheus.LogOnMessageTicketLookup(false, cacheHit)
		return database.Ticket{}, false, nil
	}

	prometheus.LogOnMessageTicketLookup(true, cacheHit)

	return ticket, true, nil
}
