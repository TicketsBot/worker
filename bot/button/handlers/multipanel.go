package handlers

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
)

type MultiPanelHandler struct{}

func (h *MultiPanelHandler) Matcher() matcher.Matcher {
	return &matcher.SimpleMatcher{
		CustomId: "multipanel",
	}
}

func (h *MultiPanelHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed),
	}
}

func (h *MultiPanelHandler) Execute(ctx *context.SelectMenuContext) {
	if len(ctx.InteractionData.Values) == 0 {
		return
	}

	panelCustomId := ctx.InteractionData.Values[0]

	panel, ok, err := dbclient.Client.Panel.GetByCustomId(ctx.GuildId(), panelCustomId)
	if err != nil {
		sentry.Error(err) // TODO: Proper context
		return
	}

	if ok {
		// TODO: Log this
		if panel.GuildId != ctx.GuildId() {
			return
		}

		// blacklist check
		blacklisted, err := dbclient.Client.Blacklist.IsBlacklisted(panel.GuildId, ctx.InteractionUser().Id)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if blacklisted {
			ctx.Reply(constants.Red, i18n.TitleBlacklisted, i18n.MessageBlacklisted)
			return
		}

		// TODO: If we have complaints of the bot sending non-ephemeral msgs in the panel channel, we can call ctx.IntoPanelContext
		_, _ = logic.OpenTicket(ctx, &panel, panel.Title)

		return
	}
}
