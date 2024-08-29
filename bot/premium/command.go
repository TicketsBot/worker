package premium

import (
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
	"strings"
)

const PremiumStoreSku uint64 = 1274473638065606656

func BuildKeyModal(guildId uint64) interaction.ModalResponseData {
	return interaction.ModalResponseData{
		CustomId: "premium_key_modal",
		Title:    i18n.GetMessageFromGuild(guildId, i18n.MessagePremiumActivateKey),
		Components: []component.Component{
			component.BuildActionRow(component.BuildInputText(component.InputText{
				Style:       component.TextStyleShort,
				CustomId:    "key",
				Label:       i18n.GetMessageFromGuild(guildId, i18n.MessagePremiumKey),
				Placeholder: utils.Ptr("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
				MinLength:   utils.Ptr(uint32(36)),
				MaxLength:   utils.Ptr(uint32(36)),
			})),
		},
	}
}

func BuildSubscriptionFoundMessage(ctx registry.CommandContext) (command.MessageResponse, error) {
	guild, err := ctx.Guild()
	if err != nil {
		return command.MessageResponse{}, err
	}

	commands, err := command.LoadCommandIds(ctx.Worker(), ctx.Worker().BotId)
	if err != nil {
		return command.MessageResponse{}, err
	}

	components := utils.Slice(component.BuildActionRow(
		component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.MessagePremiumCheckAgain),
			CustomId: "premium_check_again",
			Style:    component.ButtonStylePrimary,
			Emoji:    utils.BuildEmoji("🔎"),
		}),
		component.BuildButton(component.Button{
			Label: ctx.GetMessage(i18n.MessageJoinSupportServer),
			Style: component.ButtonStyleLink,
			Emoji: utils.BuildEmoji("❓"),
			Url:   utils.Ptr(strings.ReplaceAll(config.Conf.Bot.SupportServerInvite, "\n", "")),
		}),
	))

	embed := utils.BuildEmbed(ctx, customisation.Red, i18n.MessagePremiumSubscriptionFound, i18n.MessagePremiumSubscriptionFoundContent, nil, guild.OwnerId, commands["addadmin"], commands["viewstaff"])
	return command.NewEphemeralEmbedMessageResponseWithComponents(embed, components), nil
}

func BuildNotLinkedMessage(ctx registry.CommandContext) command.MessageResponse {
	components := utils.Slice(component.BuildActionRow(
		component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.MessagePremiumCheckAgain),
			CustomId: "premium_check_again",
			Style:    component.ButtonStylePrimary,
			Emoji:    utils.BuildEmoji("🔎"),
		}),
		component.BuildButton(component.Button{
			Label: ctx.GetMessage(i18n.MessagePremiumLinkPatreonAccount),
			Style: component.ButtonStyleLink,
			Emoji: ctx.SelectValidEmoji(customisation.EmojiPatreon, "🔗"),
			Url:   utils.Ptr("https://support.patreon.com/hc/en-us/articles/212052266-Get-my-Discord-role"), // TODO: Localised link
		}),
		component.BuildButton(component.Button{
			Label: ctx.GetMessage(i18n.MessageJoinSupportServer),
			Style: component.ButtonStyleLink,
			Emoji: utils.BuildEmoji("❓"),
			Url:   utils.Ptr(strings.ReplaceAll(config.Conf.Bot.SupportServerInvite, "\n", "")),
		}),
	))

	embed := utils.BuildEmbed(ctx, customisation.Red, i18n.TitlePremium, i18n.MessagePremiumNoSubscription, nil)
	return command.NewEphemeralEmbedMessageResponseWithComponents(embed, components)
}

func BuildDiscordNotFoundMessage(ctx registry.CommandContext) command.MessageResponse {
	embed := utils.BuildEmbed(ctx, customisation.Red, i18n.TitlePremium, i18n.MessagePremiumDiscordNoSubscription, nil)

	return command.NewEphemeralEmbedMessageResponseWithComponents(embed, utils.Slice(component.BuildActionRow(
		component.BuildButton(component.Button{
			Style: component.ButtonStylePremium,
			SkuId: utils.Ptr(PremiumStoreSku),
		}),
	)))
}
