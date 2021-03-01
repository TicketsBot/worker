package utils

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
	"time"
)

// guildId is only used to get the language
func SendEmbed(
	worker *worker.Context,
	channelId, guildId uint64, replyTo *message.MessageReference,
	colour Colour, title string, messageType translations.MessageId, fields []embed.EmbedField,
	deleteAfter int, isPremium bool,
	format ...interface{},
) {
	content := i18n.GetMessageFromGuild(guildId, messageType, format...)
	_, _ = SendEmbedWithResponse(worker, channelId, replyTo, colour, title, content, fields, deleteAfter, isPremium)
}

func SendEmbedRaw(
	worker *worker.Context,
	channel uint64, replyTo *message.MessageReference,
	colour Colour, title, content string, fields []embed.EmbedField,
	deleteAfter int, isPremium bool,
) {
	_, _ = SendEmbedWithResponse(worker, channel, replyTo, colour, title, content, fields, deleteAfter, isPremium)
}

func SendEmbedWithResponse(
	worker *worker.Context,
	channel uint64, replyTo *message.MessageReference,
	colour Colour, title, content string, fields []embed.EmbedField,
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
		self, _ := worker.Self()
		msgEmbed.SetFooter("Powered by ticketsbot.net", self.AvatarUrl(256))
	}

	data := rest.CreateMessageData{
		Embed:            msgEmbed,
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
	worker *worker.Context,
	guildId uint64,
	colour Colour, title string, messageType translations.MessageId, fields []embed.EmbedField,
	isPremium bool,
	format ...interface{},
) *embed.Embed {
	content := i18n.GetMessageFromGuild(guildId, messageType, format...)

	msgEmbed := embed.NewEmbed().
		SetColor(int(colour)).
		SetTitle(title).
		SetDescription(content)

	for _, field := range fields {
		msgEmbed.AddField(field.Name, field.Value, field.Inline)
	}

	if !isPremium {
		self, _ := worker.Self()
		msgEmbed.SetFooter("Powered by ticketsbot.net", self.AvatarUrl(256))
	}

	return msgEmbed
}

func BuildEmbedRaw(
	worker *worker.Context,
	colour Colour, title, content string, fields []embed.EmbedField,
	isPremium bool,
) *embed.Embed {
	msgEmbed := embed.NewEmbed().
		SetColor(int(colour)).
		SetTitle(title).
		SetDescription(content)

	for _, field := range fields {
		msgEmbed.AddField(field.Name, field.Value, field.Inline)
	}

	if !isPremium {
		self, _ := worker.Self()
		msgEmbed.SetFooter("Powered by ticketsbot.net", self.AvatarUrl(256))
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
