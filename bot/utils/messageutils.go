package utils

import (
	"fmt"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
	"time"
)

// guildId is only used to get the language
func SendEmbed(
	worker *worker.Context,
	channelId, guildId uint64, replyTo *message.MessageReference,
	colour constants.Colour, title string, messageType i18n.MessageId, fields []embed.EmbedField,
	deleteAfter int, isPremium bool,
	format ...interface{},
) {
	content := i18n.GetMessageFromGuild(guildId, messageType, format...)
	_, _ = SendEmbedWithResponse(worker, channelId, replyTo, colour, title, content, fields, deleteAfter, isPremium)
}

func SendEmbedRaw(
	worker *worker.Context,
	channel uint64, replyTo *message.MessageReference,
	colour constants.Colour, title, content string, fields []embed.EmbedField,
	deleteAfter int, isPremium bool,
) {
	_, _ = SendEmbedWithResponse(worker, channel, replyTo, colour, title, content, fields, deleteAfter, isPremium)
}

func SendEmbedWithResponse(
	worker *worker.Context,
	channel uint64, replyTo *message.MessageReference,
	colour constants.Colour, title, content string, fields []embed.EmbedField,
	deleteAfter int, isPremium bool,
) (message.Message, error) {
	msgEmbed := embed.NewEmbed().
		SetColor(int(colour)).
		SetTitle(title).
		SetDescription(content)

	for _, field := range fields {
		msgEmbed.AddField(field.Name, field.Value, field.Inline)
	}

	if !isPremium {
		msgEmbed.SetFooter("Powered by ticketsbot.net", "https://ticketsbot.net/assets/img/logo.png")
	}

	data := rest.CreateMessageData{
		Embeds:           []*embed.Embed{msgEmbed},
		MessageReference: replyTo,
	}

	msg, err := worker.CreateMessageComplex(channel, data)

	if err != nil {
		sentry.LogWithContext(err, errorcontext.WorkerErrorContext{
			Channel: channel,
		})

		return msg, err
	}

	if deleteAfter > 0 {
		DeleteAfter(worker, msg.ChannelId, msg.Id, deleteAfter)
	}

	return msg, err
}

func BuildEmbed(
	ctx registry.CommandContext,
	colour constants.Colour, titleId, contentId i18n.MessageId, fields []embed.EmbedField,
	format ...interface{},
) *embed.Embed {
	title := i18n.GetMessageFromGuild(ctx.GuildId(), titleId)
	content := i18n.GetMessageFromGuild(ctx.GuildId(), contentId, format...)

	msgEmbed := embed.NewEmbed().
		SetColor(int(colour)).
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
	colour constants.Colour, title, content string, fields []embed.EmbedField, tier premium.PremiumTier,
) *embed.Embed {
	msgEmbed := embed.NewEmbed().
		SetColor(int(colour)).
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

func DeleteAfter(worker *worker.Context, channelId, messageId uint64, secs int) {
	go func() {
		time.Sleep(time.Duration(secs) * time.Second)

		// Explicitly ignore error, pretty much always a 404
		// TODO: Should we log it?
		_ = worker.DeleteMessage(channelId, messageId)
	}()
}

func ReactWithCheck(worker *worker.Context, channelId, messageId uint64) {
	if err := worker.CreateReaction(channelId, messageId, "✅"); err != nil {
		sentry.LogWithContext(err, errorcontext.WorkerErrorContext{
			Channel: channelId,
		})
	}
}

func ReactWithCross(worker *worker.Context, channelId, messageId uint64) {
	if err := worker.CreateReaction(channelId, messageId, "❌"); err != nil {
		sentry.LogWithContext(err, errorcontext.WorkerErrorContext{
			Channel: channelId,
		})
	}
}

func PadDiscriminator(discrim user.Discriminator) string {
	return fmt.Sprintf("%04d", uint16(discrim))
}

func CreateReference(messageId, channelId, guildId uint64) *message.MessageReference {
	return &message.MessageReference{
		MessageId:       messageId,
		ChannelId:       channelId,
		GuildId:         guildId,
		FailIfNotExists: false,
	}
}

func CreateReferenceFromEvent(ev *events.MessageCreate) *message.MessageReference {
	return &message.MessageReference{
		MessageId:       ev.Id,
		ChannelId:       ev.ChannelId,
		GuildId:         ev.GuildId,
		FailIfNotExists: false,
	}
}

func CreateReferenceFromMessage(msg message.Message) *message.MessageReference {
	return &message.MessageReference{
		MessageId:       msg.Id,
		ChannelId:       msg.ChannelId,
		GuildId:         msg.GuildId,
		FailIfNotExists: false,
	}
}

func BlankField(inline bool) embed.EmbedField {
	return embed.EmbedField{
		Name:   "\u200b",
		Value:  "‎",
		Inline: inline,
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
