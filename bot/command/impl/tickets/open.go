package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/logic"
)

type OpenCommand struct {
}

func (OpenCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "open",
		Description:     "Opens a new ticket",
		Aliases:         []string{"new"},
		PermissionLevel: permission.Everyone,
		Category:        command.Tickets,
	}
}

func (OpenCommand) Execute(ctx command.CommandContext) {
	logic.OpenTicket(ctx.Worker, ctx.Author, ctx.GuildId, ctx.ChannelId, ctx.Id, ctx.PremiumTier > premium.None, ctx.Args, nil)
}
