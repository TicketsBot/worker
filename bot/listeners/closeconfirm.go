package listeners

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
)

func OnCloseConfirm(worker *worker.Context, data interaction.ButtonInteraction) {
	// Get whether the guild is premium
	premiumTier, err := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, worker.Token, worker.RateLimiter)
	if err != nil {
		sentry.Error(err)
		return
	}

	ctx := context.NewPanelContext(worker, data.GuildId.Value, data.ChannelId, data.Member.User.Id, premiumTier)
	logic.CloseTicket(&ctx, nil, true)
}
