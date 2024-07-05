package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
)

type PremiumCommand struct {
}

func (PremiumCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "premium",
		Description:      i18n.HelpPremium,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permission.Admin,
		Category:         command.Settings,
		DefaultEphemeral: true,
	}
}

func (c PremiumCommand) GetExecutor() interface{} {
	return c.Execute
}

func (PremiumCommand) Execute(ctx registry.CommandContext) {
	premiumTier := ctx.PremiumTier()

	// Tell user if premium is already active
	if premiumTier > premium.None {
		// Re-enable panels
		if err := dbclient.Client.Panel.EnableAll(ctx.GuildId()); err != nil {
			ctx.HandleError(err)
			return
		}

		var content i18n.MessageId
		if premiumTier == premium.Whitelabel {
			content = i18n.MessagePremiumLinkAlreadyActivatedWhitelabel
		} else {
			content = i18n.MessagePremiumLinkAlreadyActivated
		}

		ctx.ReplyWith(command.NewEphemeralEmbedMessageResponseWithComponents(
			utils.BuildEmbed(ctx, customisation.Green, i18n.TitlePremium, content, nil),
			utils.Slice(component.BuildActionRow(
				component.BuildButton(component.Button{
					Label:    ctx.GetMessage(i18n.MessagePremiumUseKeyAnyway),
					CustomId: "open_premium_key_modal",
					Style:    component.ButtonStyleSecondary,
					Emoji:    utils.BuildEmoji("🔑"),
				}),
			)),
		))

	} else {
		var patreonEmoji, keyEmoji *emoji.Emoji
		if !ctx.Worker().IsWhitelabel {
			patreonEmoji = customisation.EmojiPatreon.BuildEmoji()
			keyEmoji = utils.BuildEmoji("🔑")
		}

		// utils.EmbedField
		fields := utils.Slice(embed.EmbedField{
			Name:   ctx.GetMessage(i18n.MessagePremiumAlreadyPurchasedTitle),
			Value:  ctx.GetMessage(i18n.MessagePremiumAlreadyPurchasedDescription),
			Inline: false,
		})

		ctx.ReplyWith(command.NewEphemeralEmbedMessageResponseWithComponents(
			utils.BuildEmbed(ctx, customisation.Green, i18n.TitlePremium, i18n.MessagePremiumAbout, fields),
			utils.Slice(
				component.BuildActionRow(
					component.BuildSelectMenu(component.SelectMenu{
						CustomId: "premium_purchase_method",
						Options: utils.Slice(
							component.SelectOption{
								Label:       "Patreon", // Don't translate
								Value:       "patreon",
								Description: ctx.GetMessage(i18n.MessagePremiumMethodSelectorPatreon),
								Emoji:       patreonEmoji,
							},
							component.SelectOption{
								Label:       ctx.GetMessage(i18n.MessagePremiumGiveawayKey),
								Value:       "key",
								Description: ctx.GetMessage(i18n.MessagePremiumMethodSelectorKey),
								Emoji:       keyEmoji,
							},
						),
						Placeholder: ctx.GetMessage(i18n.MessagePremiumMethodSelector),
						Disabled:    false,
					}),
				),
				component.BuildActionRow(
					component.BuildButton(component.Button{
						Label: ctx.GetMessage(i18n.Website),
						Style: component.ButtonStyleLink,
						Emoji: utils.BuildEmoji("🔗"),
						Url:   utils.Ptr("https://ticketsbot.net/premium"),
					}),
				),
			),
		))
	}
}
