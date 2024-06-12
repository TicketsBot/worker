package button

import (
	"context"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
)

type ResponseMessage struct {
	Data command.MessageResponse
}

func (r ResponseMessage) Type() ResponseType {
	return ResponseTypeMessage
}

func (r ResponseMessage) Build() interface{} {
	return interaction.NewResponseChannelMessage(r.Data.IntoApplicationCommandData())
}

func (r ResponseMessage) HandleDeferred(interactionData interaction.MessageComponentInteraction, worker *worker.Context) error {
	_, err := rest.CreateFollowupMessage(context.Background(), interactionData.Token, worker.RateLimiter, worker.BotId, r.Data.IntoWebhookBody())
	return err
}
