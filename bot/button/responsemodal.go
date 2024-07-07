package button

import (
	"errors"
	"github.com/TicketsBot/worker"
	"github.com/rxdn/gdl/objects/interaction"
)

type ResponseModal struct {
	Data interaction.ModalResponseData
}

func (r ResponseModal) Type() ResponseType {
	return ResponseTypeModal
}

func (r ResponseModal) Build() interface{} {
	return interaction.NewModalResponse(r.Data.CustomId, r.Data.Title, r.Data.Components)
}

func (r ResponseModal) HandleDeferred(interactionData interaction.InteractionMetadata, worker *worker.Context) error {
	return errors.New("cannot defer modal response")
}
