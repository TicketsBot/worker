package messagequeue

import (
	"fmt"
	"github.com/TicketsBot/common/closerelay"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/cache"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/rest/ratelimit"
	"os"
)

// TODO: Make this good
func ListenTicketClose() {
	ch := make(chan closerelay.TicketClose)
	go closerelay.Listen(redis.Client, ch)

	for payload := range ch {
		payload := payload

		go func() {
			if payload.Reason == "" {
				payload.Reason = "No reason specified"
			}
			// Get the ticket struct
			ticket, err := dbclient.Client.Tickets.Get(payload.TicketId, payload.GuildId)
			if err != nil {
				sentry.Error(err)
				return
			}

			// Check that this is a valid ticket
			if ticket.GuildId == 0 {
				return
			}

			// Create error context for later
			errorContext := errorcontext.WorkerErrorContext{
				Guild: ticket.GuildId,
				User:  payload.UserId,
			}

			// Get bot token for guild
			var token string
			var botId uint64
			{
				whiteLabelBotId, isWhitelabel, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(payload.GuildId)
				if err != nil {
					sentry.ErrorWithContext(err, errorContext)
				}

				if isWhitelabel {
					bot, err := dbclient.Client.Whitelabel.GetByBotId(whiteLabelBotId); if err != nil {
						sentry.ErrorWithContext(err, errorContext)
						return
					}

					if bot.Token == "" {
						token = os.Getenv("WORKER_PUBLIC_TOKEN")
					} else {
						token = bot.Token
						botId = whiteLabelBotId
					}
				} else {
					token = os.Getenv("WORKER_PUBLIC_TOKEN")
				}
			}

			// Create ratelimiter
			var keyPrefix string
			if botId != 0 { // If is whitelabel
				keyPrefix = fmt.Sprintf("ratelimiter:%d", botId)
			} else {
				keyPrefix = "ratelimiter:public"
			}

			// TODO: Handle large sharding buckets - envvar?
			rateLimiter := ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis.Client, keyPrefix), 1)

			// Get whether the guild is premium for log archiver
			premiumTier, err := utils.PremiumClient.GetTierByGuildId(payload.GuildId, true, token, rateLimiter)
			if err != nil {

			}

			// Create worker context
			workerCtx := &worker.Context{
				Token:        token,
				IsWhitelabel: botId != 0,
				Cache:        cache.Client, // TODO: Less hacky
				RateLimiter:  rateLimiter,
			}

			// if ticket didnt open in the first place, no channel ID is assigned.
			// we should close it here, or it wont get closed at all
			if ticket.ChannelId == nil {
				if err := dbclient.Client.Tickets.Close(payload.TicketId, payload.GuildId); err != nil {
					sentry.ErrorWithContext(err, errorContext)
				}
				return
			}

			// ticket.ChannelId cannot be nil
			ctx := context.NewDashboardContext(workerCtx, ticket.GuildId, *ticket.ChannelId, payload.UserId, premiumTier)

			logic.CloseTicket(&ctx, &payload.Reason, true)
		}()
	}
}
