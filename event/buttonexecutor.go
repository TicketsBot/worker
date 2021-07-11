package event

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/listeners"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
)

func handleButtonPress(ctx *worker.Context, data interaction.ButtonInteraction) {
	switch data.Data.CustomId {
	case "close":
		listeners.OnCloseReact(ctx, data)
	case "close_confirm":
		listeners.OnCloseConfirm(ctx, data)
	case "claim":
		listeners.OnClaimReact(ctx, data)
	default:
		handlePanelButton(ctx, data)
	}
}

func handlePanelButton(ctx *worker.Context, data interaction.ButtonInteraction) {
	panel, ok, err := dbclient.Client.Panel.GetByCustomId(data.GuildId.Value, data.Data.CustomId)
	if err != nil {
		sentry.Error(err) // TODO: Proper context
		return
	}

	if ok {
		// TODO: Log this
		if panel.GuildId != data.GuildId.Value {
			return
		}

		// TODO: Log this
		if data.Member == nil {
			return
		}

		// get premium tier
		premiumTier := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, ctx.Token, ctx.RateLimiter)
		panelCtx := command.NewPanelContext(ctx, data.GuildId.Value, data.ChannelId, data.Member.User.Id, premiumTier)

		// blacklist check
		blacklisted, err := dbclient.Client.Blacklist.IsBlacklisted(data.GuildId.Value, data.Member.User.Id)
		if err != nil {
			panelCtx.HandleError(err)
			return
		}

		if blacklisted {
			panelCtx.Reply(utils.Red, "Blacklisted", i18n.MessageBlacklisted)
			return
		}

		logic.OpenTicket(&panelCtx, &panel, panel.Title)

		return
	}
}
