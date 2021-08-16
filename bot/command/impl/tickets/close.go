package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/rxdn/gdl/objects/interaction"
)

type CloseCommand struct {
}

func (CloseCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "close",
		Description:     i18n.HelpClose,
		PermissionLevel: permission.Everyone,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewOptionalArgument("reason", "The reason the ticket was closed", interaction.OptionTypeString, -1), // should never fail
		),
	}
}

func (c CloseCommand) GetExecutor() interface{} {
	return c.Execute
}

func (CloseCommand) Execute(ctx registry.CommandContext, reason *string) {
	logic.CloseTicket(ctx, reason, false)
}
