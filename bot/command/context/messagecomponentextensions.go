package context

import (
	"github.com/TicketsBot/worker/bot/button"
	"github.com/TicketsBot/worker/bot/command/registry"
)

type MessageComponentExtensions struct {
	ctx             registry.CommandContext
	responseChannel chan button.Response
}

func NewMessageComponentExtensions(ctx registry.CommandContext, responseChannel chan button.Response) *MessageComponentExtensions {
	return &MessageComponentExtensions{
		ctx:             ctx,
		responseChannel: responseChannel,
	}
}

func (e *MessageComponentExtensions) Modal(res button.ResponseModal) {
	e.responseChannel <- res
}
