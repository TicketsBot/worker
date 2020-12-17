package tickets

import (
	"fmt"
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
		Description:     translations.HelpClaim,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
	}
}

func (c ClaimCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ClaimCommand) Execute(ctx command.CommandContext) {
	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId()); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.Reply(utils.Red, "Error", translations.MessageNotATicketChannel)
		ctx.Reject()
		return
	}

	if err := logic.ClaimTicket(ctx.Worker(), ticket, ctx.UserId()); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyPermanent(utils.Green, "Ticket Claimed", translations.MessageClaimed, fmt.Sprintf("<@%d>", ctx.UserId()))
	ctx.Accept()
}
