package listeners

import (
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnCloseConfirm(worker *worker.Context, e *events.MessageReactionAdd) {
	// Check reaction is a ✅
	if e.UserId == worker.BotId || e.Emoji.Name != "✅" {
		return
	}

	// Verify it's the same user reacting
	if !redis.ConfirmClose(redis.Client, e.MessageId, e.UserId) {
		return
	}

	// Get whether the guild is premium
	premiumTier := utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter)

	ctx := command.NewPanelContext(worker, e.GuildId, e.ChannelId, e.UserId, premiumTier)
	logic.CloseTicket(&ctx, 0, nil, true)
}
