package event

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
)

func handleButtonPress(ctx *worker.Context, data interaction.ButtonInteraction) {
	fmt.Printf("[%d] 1\n", data.Message.Id)
	panel, ok, err := dbclient.Client.Panel.GetByCustomId(data.GuildId.Value, data.Data.CustomId)
	if err != nil {
		sentry.Error(err) // TODO: Proper context
		return
	}
	fmt.Printf("[%d] %s %v\n", data.Message.Id, panel.CustomId, ok)

	if ok {
		// TODO: Log this
		if panel.MessageId != data.Message.Id || panel.GuildId != data.GuildId.Value {
			fmt.Printf("[%d] not matching\n", data.Message.Id)
			return
		}
		fmt.Printf("[%d] matching\n", data.Message.Id)

		// TODO: Log this
		fmt.Printf("[%d] member nil: %v\n", data.Message.Id, data.Member == nil)
		if data.Member == nil {
			return
		}

		// get premium tier
		fmt.Printf("[%d] getting premium tier\n", data.Message.Id)
		premiumTier := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, ctx.Token, ctx.RateLimiter)
		fmt.Printf("[%d] got premium tier\n", data.Message.Id)
		panelCtx := command.NewPanelContext(ctx, data.GuildId.Value, data.ChannelId, data.Member.User.Id, premiumTier)

		fmt.Printf("[%d] got ctx\n", data.Message.Id)

		logic.OpenTicket(&panelCtx, &panel, panel.Title)

		return
	}
}
