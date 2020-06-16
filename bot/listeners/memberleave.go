package listeners

import (
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/autoclose"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"strings"
)

// Remove user permissions when they leave
func OnMemberLeave(worker *worker.Context, e *events.GuildMemberRemove, extra eventforwarding.Extra) {
	if err := dbclient.Client.Permissions.RemoveSupport(e.GuildId, e.User.Id); err != nil {
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
					// verify ticket exists + prevent potential panic
					if ticket.ChannelId == nil {
						return
					}

					// get self member
					self, err := worker.GetGuildMember(ticket.GuildId, worker.BotId)
					if err != nil {
						sentry.Error(err)
						return
					}

					// get premium status
					premiumTier := utils.PremiumClient.GetTierByGuildId(ticket.GuildId, true, worker.Token, worker.RateLimiter)

					logic.CloseTicket(
						worker,
						ticket.GuildId,
						*ticket.ChannelId,
						0,
						self,
						strings.Split(autoclose.AutoCloseReason, " "),
						true,
						premiumTier >= premium.Premium,
					)
				}
			}
		}
	}
}
