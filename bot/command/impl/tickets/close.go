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

func (CloseCommand) AutoCompleteHandler(data interaction.ApplicationCommandAutoCompleteInteraction, value string) []interaction.ApplicationCommandOptionChoice {
	var reasons []string
	var err error

	// If there is no text provided by the user yet, and this is a ticket channel, we can use our materialised view to
	// get the most common close reasons for that panel. Otherwise, perform a dynamic query to get the most common
	// reasons for that text for all panels.
	if len(value) > 0 {
		reasons, err = dbclient.Client.CloseReason.GetCommon(data.GuildId.Value, value, 10)
	} else {
		// Get ticket
		ticket, e := dbclient.Client.Tickets.GetByChannel(data.ChannelId)
		if e != nil {
			sentry.Error(e) // TODO: Context
			return nil
		}

		if ticket.Id == 0 {
			reasons, err = dbclient.Client.CloseReason.GetCommon(data.GuildId.Value, value, 10)
		} else {
			reasons, err = dbclient.Client.TopCloseReasonsView.Get(ticket.GuildId, ticket.PanelId)
		}
	}

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
