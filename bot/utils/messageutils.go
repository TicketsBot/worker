package utils

import (
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild/emoji"
)

func BuildEmbed(
	ctx registry.CommandContext,
	colour customisation.Colour, titleId, contentId i18n.MessageId, fields []embed.EmbedField,
	format ...interface{},
) *embed.Embed {
	title := i18n.GetMessageFromGuild(ctx.GuildId(), titleId)
	content := i18n.GetMessageFromGuild(ctx.GuildId(), contentId, format...)

	msgEmbed := embed.NewEmbed().
		SetColor(ctx.GetColour(colour)).
		SetTitle(title).
		SetDescription(content)

	for _, field := range fields {
		msgEmbed.AddField(field.Name, field.Value, field.Inline)
	}

	if ctx.PremiumTier() == premium.None {
		msgEmbed.SetFooter("Powered by ticketsbot.net", "https://ticketsbot.net/assets/img/logo.png")
	}

	return msgEmbed
}

func BuildEmbedRaw(
	colourHex int, title, content string, fields []embed.EmbedField, tier premium.PremiumTier,
) *embed.Embed {
	msgEmbed := embed.NewEmbed().
		SetColor(colourHex).
		SetTitle(title).
		SetDescription(content)

	for _, field := range fields {
		msgEmbed.AddField(field.Name, field.Value, field.Inline)
	}

	if tier == premium.None {
		msgEmbed.SetFooter("Powered by ticketsbot.net", "https://ticketsbot.net/assets/img/logo.png")
	}

	return msgEmbed
}

func GetColourForGuild(worker *worker.Context, colour customisation.Colour, guildId uint64) (int, error) {
	premiumTier, err := PremiumClient.GetTierByGuildId(guildId, true, worker.Token, worker.RateLimiter)
	if err != nil {
		return 0, err
	}

	if premiumTier > premium.None {
		colourCode, ok, err := dbclient.Client.CustomColours.Get(guildId, colour.Int16())
		if err != nil {
			return 0, err
		} else if !ok {
			return colour.Default(), nil
		} else {
			return colourCode, nil
		}
	} else {
		return colour.Default(), nil
	}
}

func EmbedFieldRaw(name, value string, inline bool) embed.EmbedField {
	return embed.EmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	}
}

func EmbedField(guildId uint64, name string, value i18n.MessageId, inline bool, format ...interface{}) embed.EmbedField {
	return embed.EmbedField{
		Name:   name,
		Value:  i18n.GetMessageFromGuild(guildId, value, format...),
		Inline: inline,
	}
}

func BuildEmoji(emote string) *emoji.Emoji {
	return &emoji.Emoji{
		Name: emote,
	}
}

func Embeds(embeds ...*embed.Embed) []*embed.Embed {
	return embeds
}
