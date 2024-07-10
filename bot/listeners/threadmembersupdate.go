package listeners

import (
	"context"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"time"
)

func OnThreadMembersUpdate(worker *worker.Context, e events.ThreadMembersUpdate) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*6) // TODO: Propagate context
	defer cancel()

	settings, err := dbclient.Client.Settings.Get(ctx, e.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
		return
	}

	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx, e.ThreadId, e.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
		return
	}

	if ticket.Id == 0 || ticket.GuildId != e.GuildId {
		return
	}

	if ticket.JoinMessageId != nil {
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

		threadStaff, err := logic.GetStaffInThread(ctx, worker, ticket, e.ThreadId)
		if err != nil {
			sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
			return
		}

		if settings.TicketNotificationChannel != nil {
			data := logic.BuildJoinThreadMessage(ctx, worker, ticket.GuildId, ticket.UserId, ticket.Id, panel, threadStaff, premiumTier)
			if _, err := worker.EditMessage(*settings.TicketNotificationChannel, *ticket.JoinMessageId, data.IntoEditMessageData()); err != nil {
				sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{Guild: e.GuildId})
			}
		}
	}
}
