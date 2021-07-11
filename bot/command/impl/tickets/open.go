package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/rxdn/gdl/objects/interaction"
)

type OpenCommand struct {
}

func (OpenCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "open",
		Description:     i18n.HelpOpen,
		Aliases:         []string{"new"},
		PermissionLevel: permission.Everyone,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewOptionalArgument("subject", "The subject of the ticket", interaction.OptionTypeString, i18n.MessageInvalidArgument), // TODO: Better invalid message
		),
	}
}

func (c OpenCommand) GetExecutor() interface{} {
	return c.Execute
}

func (OpenCommand) Execute(ctx registry.CommandContext, providedSubject *string) {
	var subject string
	if providedSubject != nil {
		subject = *providedSubject
	}

	logic.OpenTicket(ctx, nil, subject)
}
