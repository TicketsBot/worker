package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/logic"
)

type CloseCommand struct {
}

func (CloseCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "close",
		Description:     translations.HelpClose,
		PermissionLevel: permission.Everyone,
		Category:        command.Tickets,
	}
}

func (CloseCommand) Execute(ctx command.CommandContext) {
	logic.CloseTicket(ctx.Worker, ctx.GuildId, ctx.ChannelId, ctx.Id, ctx.Member, ctx.Args, false, ctx.PremiumTier > premium.None)
}
