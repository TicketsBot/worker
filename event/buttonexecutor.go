package event

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/listeners"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

// TODO: Better
// Returns whether the message may be edited
func handleButtonPress(ctx *worker.Context, data interaction.ButtonInteraction, responseCh chan registry.MessageResponse) bool {
	if strings.HasPrefix(data.Data.CustomId, "rate_") {
		go listeners.OnRate(ctx, data)
		return false
	} else if strings.HasPrefix(data.Data.CustomId, "viewstaff_") {
		go listeners.OnViewStaffClick(ctx, data, responseCh)
		return true
	} else {
		if data.GuildId.Value == 0 {
			return false
		}

		switch data.Data.CustomId {
		case "close":
			go listeners.OnCloseReact(ctx, data)
		case "close_confirm":
			go listeners.OnCloseConfirm(ctx, data)
		case "claim":
			go listeners.OnClaimReact(ctx, data)
		default:
			go handlePanelButton(ctx, data)
		}

		return false
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
		premiumTier, err := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, ctx.Token, ctx.RateLimiter)
		if err != nil {
			sentry.Error(err)
			return
		}

		panelCtx := context.NewPanelContext(ctx, data.GuildId.Value, data.ChannelId, data.Member.User.Id, premiumTier)

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
