package setup

import (
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/rxdn/gdl/objects/channel/message"
)

type Stage interface {
	State() State
	Prompt() translations.MessageId
	Default() string
	Process(worker *worker.Context, msg message.Message)
}
