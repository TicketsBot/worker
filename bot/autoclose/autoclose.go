package autoclose

import (
	"github.com/TicketsBot/common/autoclose"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/cache"
	"strings"
)

var reason = "Automatically closed due to inactivity"

func ListenAutoClose(cache *cache.PgCache) {
	ch := make(chan autoclose.Ticket)
	go autoclose.Listen(redis.Client, ch)

	for ticket := range ch {
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
				sentry.Error(err)
				return
			}

			// get premium status
			premiumTier := utils.PremiumClient.GetTierByGuildId(ticket.GuildId, true, worker.Token, worker.RateLimiter)

			logic.CloseTicket(worker, ticket.GuildId, *ticket.ChannelId, 0, self, strings.Split(reason, " "), true, premiumTier >= premium.Premium)
		}()
	}
}
