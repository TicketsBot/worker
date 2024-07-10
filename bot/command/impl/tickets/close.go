package tickets

import (
	"context"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
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
		Timeout: constants.TimeoutCloseTicket,
	}
}

func (c CloseCommand) GetExecutor() interface{} {
	return c.Execute
}

func (CloseCommand) Execute(ctx registry.CommandContext, reason *string) {
	logic.CloseTicket(ctx, ctx, reason, false)
}

func (CloseCommand) AutoCompleteHandler(data interaction.ApplicationCommandAutoCompleteInteraction, value string) []interaction.ApplicationCommandOptionChoice {
	var reasons []string
	var err error

	// Get ticket
	ticket, e := dbclient.Client.Tickets.GetByChannelAndGuild(context.Background(), data.ChannelId, data.GuildId.Value)
	if e != nil {
		sentry.Error(e) // TODO: Context
		return nil
	}

	ctx, cancel := utils.ContextTimeout(time.Millisecond * 1500)
	defer cancel()

	// If there is no text provided by the user yet, and this is a ticket channel, we can use our materialised view to
	// get the most common close reasons for that panel. Otherwise, perform a dynamic query to get the most common
	// reasons for that text for all panels.
	if len(value) == 0 {
		var panelId *int
		if ticket.Id != 0 {
			panelId = ticket.PanelId
		}

		reasons, err = dbclient.Analytics.GetTopCloseReasons(ctx, data.GuildId.Value, panelId)
	} else {
		var panelId *int
		if ticket.Id != 0 {
			panelId = ticket.PanelId
		}

		reasons, err = dbclient.Analytics.GetTopCloseReasonsWithPrefix(ctx, data.GuildId.Value, panelId, value)
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
