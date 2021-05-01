package listeners

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"time"
)

func OnCloseReact(worker *worker.Context, e *events.MessageReactionAdd) {
	// Check the right emoji has been used
	if e.Emoji.Name != "🔒" {
		return
	}

	// Create error context for later
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   e.GuildId,
		User:    e.UserId,
		Channel: e.ChannelId,
		Shard:   worker.ShardId,
	}

	// In DMs
	if e.GuildId == 0 {
		return
	}

	// Get user object
	user, err := worker.GetUser(e.UserId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Ensure that the user is an actual user, not a bot
	if user.Bot {
		return
	}

	// Get the ticket properties
	ticket, err := dbclient.Client.Tickets.GetByChannel(e.ChannelId); if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Check that this channel is a ticket channel
	if ticket.GuildId == 0 {
		return
	}

	// Check that the ticket has a welcome message
	if ticket.WelcomeMessageId == nil {
		return
	}

	// Check that the message being reacted to is the welcome message
	if e.MessageId != *ticket.WelcomeMessageId {
		return
	}

	closeConfirmation, err := dbclient.Client.CloseConfirmation.Get(e.GuildId); if err != nil {
		sentry.LogWithContext(err, errorContext)
		return
	}

	// Get whether the guild is premium
	premiumTier := utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter)

	if closeConfirmation {
		// Remove reaction
		_ = worker.DeleteUserReaction(e.ChannelId, e.MessageId, e.UserId, e.Emoji.Name) // Error is probably a 403, we can ignore

		// Make sure user can close;
		// Get user's permissions level
		permissionLevel, err := permission.GetPermissionLevel(utils.ToRetriever(worker), *e.Member, e.GuildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		if permissionLevel == permission.Everyone {
			usersCanClose, err := dbclient.Client.UsersCanClose.Get(e.GuildId); if err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}

			if (permissionLevel == permission.Everyone && ticket.UserId != e.UserId) || (permissionLevel == permission.Everyone && !usersCanClose) {
				utils.SendEmbed(worker, e.ChannelId, e.GuildId, nil, utils.Red, "Error", translations.MessageCloseNoPermission, nil, 30, premiumTier > premium.None)
				return
			}
		}

		// Send confirmation message
		msg, err := utils.SendEmbedWithResponse(worker, e.ChannelId, nil, utils.Green, "Close Confirmation", "React with ✅ to confirm you want to close the ticket", nil, 10, premiumTier > premium.None)
		if err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}

		if err := redis.SetCloseConfirmation(redis.Client, msg.Id, e.UserId); err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}

		time.Sleep(250 * time.Millisecond)

		// Add reaction - error likely 403
		if err = worker.CreateReaction(e.ChannelId, msg.Id, "✅"); err != nil {
			sentry.LogWithContext(err, errorContext)
		}
	} else {
		// No need to remove the reaction since we're deleting the channel anyway
		ctx := command.NewPanelContext(worker, e.GuildId, e.ChannelId, e.UserId, premiumTier)
		logic.CloseTicket(&ctx, 0, nil, true)
	}
}

