package impl

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/logic"
)

type CloseCommand struct {
}

func (CloseCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "close",
		Description:     "Closes the current ticket",
		PermissionLevel: permission.Everyone,
		Category:        command.Tickets,
	}
}

func (CloseCommand) Execute(ctx command.CommandContext) {
	logic.CloseTicket(ctx.Worker, ctx.GuildId, ctx.ChannelId, ctx.Id, ctx.Member, ctx.Args, false, ctx.PremiumTier > premium.None)
}
