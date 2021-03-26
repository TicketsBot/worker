package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/rxdn/gdl/objects/interaction"
)

type CloseCommand struct {
}

func (CloseCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "close",
		Description:     translations.HelpClose,
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
	member, err := ctx.Member()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	logic.CloseTicket(ctx.Worker(), ctx.GuildId(), ctx.ChannelId(), 0, member, reason, false, ctx.PremiumTier() > premium.None)
}
