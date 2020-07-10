package tickets

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/user"
)

type TransferCommand struct {
}

func (TransferCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "transfer",
		Description:     translations.HelpTransfer,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
	}
}

func (TransferCommand) Execute(ctx command.CommandContext) {
	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageNotATicketChannel)
		ctx.ReactWithCross()
		return
	}

	target, found := ctx.GetMentionedStaff()
	if !found {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageInvalidUser)
		ctx.ReactWithCross()
		return
	}

	if err := logic.ClaimTicket(ctx.Worker, ticket, target); err != nil {
		ctx.HandleError(err)
		return
	}

	var mention string
	{
		u := user.User{Id: target}
		mention = u.Mention()
	}

	ctx.SendEmbedNoDelete(utils.Green, "Ticket Claimed", translations.MessageClaimed, mention)
	ctx.ReactWithCheck()
}
