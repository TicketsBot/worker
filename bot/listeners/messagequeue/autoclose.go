package messagequeue

import (
	"github.com/TicketsBot/common/autoclose"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	gdlUtils "github.com/rxdn/gdl/utils"
)

const AutoCloseReason = "Automatically closed due to inactivity"

func ListenAutoClose() {
	ch := make(chan autoclose.Ticket)
	go autoclose.Listen(redis.Client, ch)

	for ticket := range ch {
		statsd.Client.IncrementKey(statsd.AutoClose)

		ticket := ticket
		go func() {
			// get worker
			worker, err := buildContext(ticket, cache.Client)
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

			// query already checks, but just to be sure
			if ticket.ChannelId == nil {
				return
			}

			// get premium status
			premiumTier, err := utils.PremiumClient.GetTierByGuildId(ticket.GuildId, true, worker.Token, worker.RateLimiter)
			if err != nil {
				sentry.Error(err)
				return
			}

			ctx := command.NewAutoCloseContext(worker, ticket.GuildId, *ticket.ChannelId, worker.BotId, premiumTier)
			logic.CloseTicket(&ctx, gdlUtils.StrPtr(AutoCloseReason), true)
		}()
	}
}
