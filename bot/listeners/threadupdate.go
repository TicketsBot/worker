package listeners

import (
	"context"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	cmdcontext "github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"time"
)

func OnThreadUpdate(worker *worker.Context, e events.ThreadUpdate) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*6) // TODO: Propagate context
	defer cancel()

	if e.ThreadMetadata == nil {
		return
	}

	settings, err := dbclient.Client.Settings.Get(ctx, e.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
		return
	}

	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx, e.Id, e.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
		return
	}

	if ticket.Id == 0 || ticket.GuildId != e.GuildId {
		return
	}

	var panel *database.Panel
	if ticket.PanelId != nil {
		tmp, err := dbclient.Client.Panel.GetById(ctx, *ticket.PanelId)
		if err != nil {
			sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
			return
		}

		if tmp.PanelId != 0 && e.GuildId == tmp.GuildId {
			panel = &tmp
		}
	}

	premiumTier, err := utils.PremiumClient.GetTierByGuildId(ctx, e.GuildId, true, worker.Token, worker.RateLimiter)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
		return
	}

	// Handle thread being unarchived
	if !ticket.Open && !e.ThreadMetadata.Archived {
		if err := dbclient.Client.Tickets.SetOpen(ctx, ticket.GuildId, ticket.Id); err != nil {
			sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
			return
		}

		if settings.TicketNotificationChannel != nil {
			staffCount, err := logic.GetStaffInThread(ctx, worker, ticket, e.Id)
			if err != nil {
				sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
				return
			}

			data := logic.BuildThreadReopenMessage(ctx, worker, ticket.GuildId, ticket.UserId, ticket.Id, panel, staffCount, premiumTier)
			msg, err := worker.CreateMessageComplex(*settings.TicketNotificationChannel, data.IntoCreateMessageData())
			if err != nil {
				sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
				return
			}

			if err := dbclient.Client.Tickets.SetJoinMessageId(ctx, ticket.GuildId, ticket.Id, &msg.Id); err != nil {
				sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
				return
			}
		}
	} else if ticket.Open && e.ThreadMetadata.Archived { // Handle ticket being archived on its own
		ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutCloseTicket)
		defer cancel()

		cc := cmdcontext.NewAutoCloseContext(ctx, worker, ticket.GuildId, e.Id, worker.BotId, premiumTier)
		logic.CloseTicket(ctx, cc, utils.Ptr("Thread was archived"), true) // TODO: Translate
	}
}
