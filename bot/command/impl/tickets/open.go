package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
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
	settings, err := dbclient.Client.Settings.Get(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if settings.DisableOpenCommand {
		return
	}

	var subject string
	if providedSubject != nil {
		subject = *providedSubject
	}

	logic.OpenTicket(ctx, nil, subject)
}
