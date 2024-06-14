package listeners

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnThreadUpdate(worker *worker.Context, e events.ThreadUpdate) {
	if e.ThreadMetadata == nil {
		return
	}

	settings, err := dbclient.Client.Settings.Get(e.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
		return
	}

	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(worker.Context, e.Id, e.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
		return
	}

	if ticket.Id == 0 || ticket.GuildId != e.GuildId {
		return
	}

	var panel *database.Panel
	if ticket.PanelId != nil {
		tmp, err := dbclient.Client.Panel.GetById(*ticket.PanelId)
		if err != nil {
			sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
			return
		}

		if tmp.PanelId != 0 && e.GuildId == tmp.GuildId {
			panel = &tmp
		}
	}

	premiumTier, err := utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
		return
	}

	// Handle thread being unarchived
	if !ticket.Open && !e.ThreadMetadata.Archived {
		if err := dbclient.Client.Tickets.SetOpen(ticket.GuildId, ticket.Id); err != nil {
			sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
			return
		}

		if settings.TicketNotificationChannel != nil {
			staffCount, err := logic.GetStaffInThread(worker, ticket, e.Id)
			if err != nil {
				sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
				return
			}

			data := logic.BuildThreadReopenMessage(worker, ticket.GuildId, ticket.UserId, ticket.Id, panel, staffCount, premiumTier)
			msg, err := worker.CreateMessageComplex(*settings.TicketNotificationChannel, data.IntoCreateMessageData())
			if err != nil {
				sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
				return
			}

			if err := dbclient.Client.Tickets.SetJoinMessageId(ticket.GuildId, ticket.Id, &msg.Id); err != nil {
				sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
				return
			}
		}
	} else if ticket.Open && e.ThreadMetadata.Archived { // Handle ticket being archived on its own
		ctx := context.NewAutoCloseContext(worker, ticket.GuildId, e.Id, worker.BotId, premiumTier)
		logic.CloseTicket(ctx, utils.Ptr("Thread was archived"), true) // TODO: Translate
	}
}
