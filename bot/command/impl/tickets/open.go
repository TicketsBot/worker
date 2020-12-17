package tickets

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/rxdn/gdl/objects/interaction"
)

type OpenCommand struct {
}

func (OpenCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "open",
		Description:     translations.HelpOpen,
		Aliases:         []string{"new"},
		PermissionLevel: permission.Everyone,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewOptionalArgument("subject", "The subject of the ticket", interaction.OptionTypeString, translations.MessageInvalidArgument), // TODO: Better invalid message
		),
	}
}

func (c OpenCommand) GetExecutor() interface{} {
	return c.Execute
}

func (OpenCommand) Execute(ctx command.CommandContext, providedSubject *string) {
	var subject string
	if providedSubject != nil {
		subject = *providedSubject
	}

	logic.OpenTicket(ctx, nil, subject)
}
