package registry

import (
	"github.com/TicketsBot/worker"
	"github.com/rxdn/gdl/objects/interaction"
)

type ButtonHandler interface {
	Matches(customId string) bool
	Properties() Properties
	Execute(ctx *worker.Context, data interaction.ButtonInteraction)
}

type Properties struct {
	DMsAllowed bool
}
