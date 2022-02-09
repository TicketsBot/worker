package button

import (
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
)

type ResponseEdit struct {
	Data command.MessageResponse
}

func (r ResponseEdit) Type() ResponseType {
	return ResponseTypeEdit
}

func (r ResponseEdit) Build() interface{} {
	return interaction.NewResponseUpdateMessage(r.Data.IntoUpdateMessageResponse())
}

func (r ResponseEdit) HandleDeferred(interactionData interaction.MessageComponentInteraction, worker *worker.Context) error {
	_, err := rest.EditOriginalInteractionResponse(interactionData.Token, worker.RateLimiter, interactionData.ApplicationId, r.Data.IntoWebhookEditBody())
	return err
}
