package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/customisation"
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
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"new"},
		PermissionLevel: permission.Everyone,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewOptionalArgument("subject", "The subject of the ticket", interaction.OptionTypeString, "infallible"),
		),
		DefaultEphemeral: true,
		Timeout:          constants.TimeoutOpenTicket,
	}
}

func (c OpenCommand) GetExecutor() interface{} {
	return c.Execute
}

func (OpenCommand) Execute(ctx *context.SlashCommandContext, providedSubject *string) {
	settings, err := ctx.Settings()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if settings.DisableOpenCommand {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageOpenCommandDisabled)
		return
	}

	var subject string
	if providedSubject != nil {
		subject = *providedSubject
	}

	logic.OpenTicket(ctx.Context, ctx, nil, subject, nil)
}
