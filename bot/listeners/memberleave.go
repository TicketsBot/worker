package listeners

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/listeners/messagequeue"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	gdlUtils "github.com/rxdn/gdl/utils"
)

// Remove user permissions when they leave
func OnMemberLeave(worker *worker.Context, e events.GuildMemberRemove) {
	if err := dbclient.Client.Permissions.RemoveSupport(e.GuildId, e.User.Id); err != nil {
		sentry.Error(err)
	}

	if err := utils.ToRetriever(worker).Cache().DeleteCachedPermissionLevel(e.GuildId, e.User.Id); err != nil {
		sentry.Error(err)
	}

	// auto close
	settings, err := dbclient.Client.AutoClose.Get(e.GuildId)
	if err != nil {
		sentry.Error(err)
	} else {
		// check setting is enabled
		if settings.Enabled && settings.OnUserLeave != nil && *settings.OnUserLeave {
			// get open tickets by user
			tickets, err := dbclient.Client.Tickets.GetOpenByUser(e.GuildId, e.User.Id)
			if err != nil {
				sentry.Error(err)
			} else {
				for _, ticket := range tickets {
					isExcluded, err := dbclient.Client.AutoCloseExclude.IsExcluded(e.GuildId, ticket.Id)
					if err != nil {
						sentry.Error(err)
						continue
					}

					if isExcluded {
						continue
					}

					// verify ticket exists + prevent potential panic
					if ticket.ChannelId == nil {
						return
					}

					// get premium status
					premiumTier, err := utils.PremiumClient.GetTierByGuildId(ticket.GuildId, true, worker.Token, worker.RateLimiter)
					if err != nil {
						sentry.Error(err)
						return
					}

					ctx := context.NewAutoCloseContext(worker, e.GuildId, *ticket.ChannelId, worker.BotId, premiumTier)
					logic.CloseTicket(ctx, gdlUtils.StrPtr(messagequeue.AutoCloseReason), true)
				}
			}
		}
	}
}
