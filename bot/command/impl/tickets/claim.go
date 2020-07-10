package tickets

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
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
		ctx.SendEmbed(utils.Red, "Error", translations.MessageNotATicketChannel)
		ctx.ReactWithCross()
		return
	}

	if err := logic.ClaimTicket(ctx.Worker, ticket, ctx.Author.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbedNoDelete(utils.Green, "Ticket Claimed", translations.MessageClaimed, ctx.Author.Mention())
	ctx.ReactWithCheck()
}
