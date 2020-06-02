package listeners

import (
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnCloseConfirm(worker *worker.Context, e *events.MessageReactionAdd, extra eventforwarding.Extra) {
	// Check reaction is a ✅
	if e.UserId == worker.BotId || e.Emoji.Name != "✅" {
		return
	}

	// Verify it's the same user reacting
	if !redis.ConfirmClose(redis.Client, e.MessageId, e.UserId) {
		return
	}

	// Create error context for later
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   e.GuildId,
		User:    e.UserId,
		Channel: e.ChannelId,
		Shard:   worker.ShardId,
	}

	// Get whether the guild is premium
	isPremium := utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter) > premium.None

	// Get the member object
	member, err := worker.GetGuildMember(e.GuildId, e.UserId)
	if err != nil {
		sentry.LogWithContext(err, errorContext)
		return
	}

	logic.CloseTicket(worker, e.GuildId, e.ChannelId, 0, member, nil, true, isPremium)
}
