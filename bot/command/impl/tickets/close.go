package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type CloseCommand struct {
}

func (c CloseCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "close",
		Description:     i18n.HelpClose,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewOptionalAutocompleteableArgument("reason", "The reason the ticket was closed", interaction.OptionTypeString, "infallible", c.AutoCompleteHandler), // should never fail
		),
	}
}

func (c CloseCommand) GetExecutor() interface{} {
	return c.Execute
}

func (CloseCommand) Execute(ctx registry.CommandContext, reason *string) {
	logic.CloseTicket(ctx, reason)
}

// AutoCompleteHandler TODO: Per panel
func (CloseCommand) AutoCompleteHandler(data interaction.ApplicationCommandAutoCompleteInteraction, value string) []interaction.ApplicationCommandOptionChoice {
	reasons, err := dbclient.Client.CloseReason.GetCommon(data.GuildId.Value, value, 10)
	if err != nil {
		sentry.Error(err) // TODO: Context
		return nil
	}

	choices := make([]interaction.ApplicationCommandOptionChoice, len(reasons))
	for i, reason := range reasons {
		choices[i] = utils.StringChoice(reason)
	}

	return choices
}
