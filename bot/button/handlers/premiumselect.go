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
	prem "github.com/TicketsBot/worker/bot/premium"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction/component"
	"strings"
)

type PremiumKeyOpenHandler struct{}

func (h *PremiumKeyOpenHandler) Matcher() matcher.Matcher {
	return matcher.NewSimpleMatcher("premium_purchase_method")
}

func (h *PremiumKeyOpenHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed, registry.CanEdit),
	}
}

func (h *PremiumKeyOpenHandler) Execute(ctx *context.SelectMenuContext) {
	permLevel, err := ctx.UserPermissionLevel()
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
		tier, err := utils.PremiumClient.GetTierByUser(ctx.UserId(), false)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if tier == premium.None {
			var patreonEmoji *emoji.Emoji
			if ctx.Worker().IsWhitelabel {
				patreonEmoji = utils.BuildEmoji("🔗")
			} else {
				patreonEmoji = utils.EmojiPatreon.BuildEmoji()
			}

			components := utils.Slice(component.BuildActionRow(
				component.BuildButton(component.Button{
					Label: ctx.GetMessage(i18n.MessagePremiumLinkPatreonAccount),
					Style: component.ButtonStyleLink,
					Emoji: patreonEmoji,
					Url:   utils.Ptr("https://support.patreon.com/hc/en-us/articles/212052266-Get-my-Discord-role"), // TODO: Localised link
				}),
				component.BuildButton(component.Button{
					Label: ctx.GetMessage(i18n.MessageJoinSupportServer),
					Style: component.ButtonStyleLink,
					Emoji: utils.BuildEmoji("❓"),
					Url:   utils.Ptr(strings.ReplaceAll(config.Conf.Bot.SupportServerInvite, "\n", "")),
				}),
			))

			ctx.EditWithComponents(customisation.Red, i18n.TitlePremium, i18n.MessagePremiumNoSubscription, components)
		} else {
			res, err := prem.BuildSubscriptionFoundMessage(ctx)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			ctx.Edit(res)
		}
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
