package handlers

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/button"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	prem "github.com/TicketsBot/worker/bot/premium"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction/component"
	"time"
)

type PremiumKeyOpenHandler struct{}

func (h *PremiumKeyOpenHandler) Matcher() matcher.Matcher {
	return matcher.NewSimpleMatcher("premium_purchase_method")
}

func (h *PremiumKeyOpenHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.GuildAllowed, registry.CanEdit),
		Timeout: time.Second * 5,
	}
}

func (h *PremiumKeyOpenHandler) Execute(ctx *context.SelectMenuContext) {
	permLevel, err := ctx.UserPermissionLevel(ctx)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permLevel < permission.Admin {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNoPermission)
		return
	}

	if len(ctx.InteractionData.Values) == 0 {
		return
	}

	option := ctx.InteractionData.Values[0]
	if option == "patreon" {
		entitlement, err := dbclient.Client.LegacyPremiumEntitlements.GetUserTier(ctx, ctx.UserId(), premium.PatreonGracePeriod)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if entitlement == nil {
			ctx.Edit(prem.BuildPatreonNotLinkedMessage(ctx))
		} else {
			res, err := prem.BuildPatreonSubscriptionFoundMessage(ctx, entitlement)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			ctx.Edit(res)
		}
	} else if option == "discord" {
		ctx.Edit(prem.BuildDiscordNotFoundMessage(ctx))
	} else if option == "key" {
		ctx.Modal(button.ResponseModal{
			Data: prem.BuildKeyModal(ctx.GuildId()),
		})

		components := utils.Slice(component.BuildActionRow(component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.MessagePremiumOpenForm),
			CustomId: "open_premium_key_modal",
			Style:    component.ButtonStylePrimary,
			Emoji:    utils.BuildEmoji("🔑"),
		})))

		ctx.EditWithComponents(customisation.Green, i18n.TitlePremium, i18n.MessagePremiumOpenFormDescription, components)
	} else {
		ctx.HandleError(fmt.Errorf("Invalid premium purchase method: %s", option))
		return
	}
}
