package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/rxdn/gdl/objects/interaction"
)

type CloseCommand struct {
}

func (CloseCommand) Properties() command.Properties {
	return command.Properties{
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

func (CloseCommand) Execute(ctx command.CommandContext, reason *string) {
	logic.CloseTicket(ctx.Worker(), ctx.GuildId(), ctx.ChannelId(), ctx.Id, ctx.Member, reason, false, ctx.PremiumTier > premium.None)
}
