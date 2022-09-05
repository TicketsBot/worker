package handlers

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	prem "github.com/TicketsBot/worker/bot/premium"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
)

type PremiumCheckAgain struct{}

func (h *PremiumCheckAgain) Matcher() matcher.Matcher {
	return &matcher.SimpleMatcher{
		CustomId: "premium_check_again",
	}
}

func (h *PremiumCheckAgain) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed),
	}
}

func (h *PremiumCheckAgain) Execute(ctx *context.ButtonContext) {
	// Get permission level
	permissionLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permissionLevel < permission.Admin {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNoPermission)
		return
	}

	ctx.EditWith(customisation.Green, i18n.MessagePremiumChecking, i18n.MessagePremiumPleaseWait)

	if err := utils.PremiumClient.DeleteCachedTier(ctx.GuildId()); err != nil {
		ctx.HandleError(err)
		return
	}

	if ctx.PremiumTier() > premium.None {
		ctx.EditWith(customisation.Green, i18n.Success, i18n.MessagePremiumSuccessAfterCheck)
	} else {
		res, err := prem.BuildSubscriptionFoundMessage(ctx)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		ctx.Edit(res)
	}
}
