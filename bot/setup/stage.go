package setup

import (
	"github.com/TicketsBot/worker"
	"github.com/rxdn/gdl/objects/channel/message"
)

type Stage interface {
	State() State
	Prompt() string
	Default() string
	Process(worker *worker.Context, msg message.Message)
}
