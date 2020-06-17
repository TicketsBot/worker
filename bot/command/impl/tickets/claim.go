package tickets

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
)

type ClaimCommand struct {
}

func (ClaimCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "claim",
		Description:     "Assigns a single staff member to a ticket",
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
	}
}

func (ClaimCommand) Execute(ctx command.CommandContext) {
	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.SendEmbed(utils.Red, "Error", "This is not a ticket channel")
		ctx.ReactWithCross()
		return
	}

	if err := logic.ClaimTicket(ctx.Worker, ticket, ctx.Author.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbedNoDelete(utils.Green, "Ticket Claimed", fmt.Sprintf("Your ticket will be handled by %s", ctx.Author.Mention()))
	ctx.ReactWithCheck()
}
