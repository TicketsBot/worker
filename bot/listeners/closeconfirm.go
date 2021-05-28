package listeners

import (
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
)

func OnCloseConfirm(worker *worker.Context, data interaction.ButtonInteraction) {
	// Get whether the guild is premium
	premiumTier := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, worker.Token, worker.RateLimiter)

	ctx := command.NewPanelContext(worker, data.GuildId.Value, data.ChannelId, data.Member.User.Id, premiumTier)
	logic.CloseTicket(&ctx, 0, nil, true)
}
