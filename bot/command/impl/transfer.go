package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
)

type TransferCommand struct {
}

func (TransferCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "transfer",
		Description:     "Transfers a claimed ticket to another user",
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
	}
}

func (TransferCommand) Execute(ctx command.CommandContext) {
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

	target, found := ctx.GetMentionedStaff()
	if !found {
		ctx.SendEmbed(utils.Red, "Error", "Couldn't find the target user")
		ctx.ReactWithCross()
		return
	}

	if err := logic.ClaimTicket(ctx.Worker, ticket, target); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbedNoDelete(utils.Green, "Ticket Claimed", fmt.Sprintf("Your ticket will be handled by <@%d>", target))
	ctx.ReactWithCheck()
}
