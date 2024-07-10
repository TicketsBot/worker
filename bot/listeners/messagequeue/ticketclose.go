package messagequeue

import (
	"context"
	"github.com/TicketsBot/common/closerelay"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/cache"
	cmdcontext "github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
)

// TODO: Make this good
func ListenTicketClose() {
	ch := make(chan closerelay.TicketClose)
	go closerelay.Listen(redis.Client, ch)

	for payload := range ch {
		payload := payload

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutCloseTicket)
			defer cancel()

			if payload.Reason == "" {
				payload.Reason = "No reason specified"
			}

			// Get the ticket struct
			ticket, err := dbclient.Client.Tickets.Get(ctx, payload.TicketId, payload.GuildId)
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
				whiteLabelBotId, isWhitelabel, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(ctx, payload.GuildId)
				if err != nil {
					sentry.ErrorWithContext(err, errorContext)
				}

				if isWhitelabel {
					bot, err := dbclient.Client.Whitelabel.GetByBotId(ctx, whiteLabelBotId)
					if err != nil {
						sentry.ErrorWithContext(err, errorContext)
						return
					}

					if bot.Token == "" {
						token = config.Conf.Discord.Token
					} else {
						token = bot.Token
						botId = whiteLabelBotId
					}
				} else {
					token = config.Conf.Discord.Token
				}
			}

			// Create worker context
			workerCtx := &worker.Context{
				Token:        token,
				IsWhitelabel: botId != 0,
				Cache:        cache.Client, // TODO: Less hacky
				RateLimiter:  nil,          // Use http-proxy ratelimit functionality
			}

			// Get whether the guild is premium for log archiver
			premiumTier, err := utils.PremiumClient.GetTierByGuildId(ctx, payload.GuildId, true, token, workerCtx.RateLimiter)
			if err != nil {
				sentry.ErrorWithContext(err, errorContext)
				return
			}

			// if ticket didnt open in the first place, no channel ID is assigned.
			// we should close it here, or it wont get closed at all
			if ticket.ChannelId == nil {
				if err := dbclient.Client.Tickets.Close(ctx, payload.TicketId, payload.GuildId); err != nil {
					sentry.ErrorWithContext(err, errorContext)
				}
				return
			}

			// ticket.ChannelId cannot be nil
			cc := cmdcontext.NewDashboardContext(ctx, workerCtx, ticket.GuildId, *ticket.ChannelId, payload.UserId, premiumTier)
			logic.CloseTicket(ctx, &cc, &payload.Reason, false)
		}()
	}
}
