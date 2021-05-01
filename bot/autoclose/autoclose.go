package autoclose

/*import (
	"github.com/TicketsBot/common/autoclose"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/rest/request"
)*/

const AutoCloseReason = "Automatically closed due to inactivity"

/*func ListenAutoClose(cache *cache.PgCache) {
	ch := make(chan autoclose.Ticket)
	go autoclose.Listen(redis.Client, ch)

	for ticket := range ch {
		ticket := ticket
		go func() {
			// get worker
			worker, err := buildContext(ticket, cache)
			if err != nil {
				sentry.Error(err)
				return
			}

			// get ticket
			ticket, err := dbclient.Client.Tickets.Get(ticket.TicketId, ticket.GuildId)
			if err != nil {
				sentry.Error(err)
				return
			}

			// verify ticket exists + prevent potential panic
			if ticket.ChannelId == nil {
				return
			}

			// get self member
			self, err := worker.GetGuildMember(ticket.GuildId, worker.BotId)
			if err != nil {
				// We are no longer in the guild and can exclude all tickets
				if restError, ok := err.(request.RestError); ok && restError.StatusCode == 403 {
					if err := dbclient.Client.AutoCloseExclude.ExcludeAll(ticket.GuildId); err != nil {
						sentry.Error(err)
					}
				}

				sentry.Error(err)
				return
			}

			// get premium status
			premiumTier := utils.PremiumClient.GetTierByGuildId(ticket.GuildId, true, worker.Token, worker.RateLimiter)

			logic.CloseTicket(worker, ticket.GuildId, *ticket.ChannelId, 0, self, &AutoCloseReason, true, premiumTier >= premium.Premium)
		}()
	}
}*/
