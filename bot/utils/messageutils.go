package utils

import (
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/rest"
	"time"
)

type SentMessage struct {
	Worker  *worker.Context
	Message *message.Message
}

func SendEmbed(worker *worker.Context, channel uint64, colour Colour, title, content string, fields []embed.EmbedField, deleteAfter int, isPremium bool) {
	_, _ = SendEmbedWithResponse(worker, channel, colour, title, content, fields, deleteAfter, isPremium)
}

func SendEmbedWithResponse(worker *worker.Context, channel uint64, colour Colour, title, content string, fields []embed.EmbedField, deleteAfter int, isPremium bool) (message.Message, error) {
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

	// Explicitly ignore error because it's usually a 403 (missing permissions)
	msg, err := worker.CreateMessageComplex(channel, rest.CreateMessageData{
		Embed: msgEmbed,
	})

	if err != nil {
		sentry.LogWithContext(err, sentry.ErrorContext{
			Channel: channel,
		})

		return msg, err
	}

	if deleteAfter > 0 {
		DeleteAfter(SentMessage{worker, &msg}, deleteAfter)
	}

	return msg, err
}

func DeleteAfter(msg SentMessage, secs int) {
	go func() {
		time.Sleep(time.Duration(secs) * time.Second)

		// Fix a panic
		if msg.Message != nil && msg.Worker != nil {
			// Explicitly ignore error, pretty much always a 404
			_ = msg.Worker.DeleteMessage(msg.Message.ChannelId, msg.Message.Id)
		}
	}()
}

func ReactWithCheck(worker *worker.Context, channelId, messageId uint64) {
	if err := worker.CreateReaction(channelId, messageId, "✅"); err != nil {
		sentry.LogWithContext(err, sentry.ErrorContext{
			Channel: channelId,
		})
	}
}

func ReactWithCross(worker *worker.Context, channelId, messageId uint64) {
	if err := worker.CreateReaction(channelId, messageId, "❌"); err != nil {
		sentry.LogWithContext(err, sentry.ErrorContext{
			Channel: channelId,
		})
	}
}

func PadDiscriminator(discrim uint16) string {
	return fmt.Sprintf("%04d", discrim)
}
