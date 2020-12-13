package tickets

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
)

type TransferCommand struct {
}

func (TransferCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "transfer",
		Description:     translations.HelpTransfer,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "Support representative to transfer the ticket to", interaction.OptionTypeUser, translations.MessageInvalidUser),
		),
	}
}

func (c TransferCommand) GetExecutor() interface{} {
	return c.Execute
}

func (TransferCommand) Execute(ctx command.CommandContext, userId uint64) {
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

	member, err := ctx.Worker.GetGuildMember(ctx.GuildId, userId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	permissionLevel := permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker), member, ctx.GuildId)
	if permissionLevel < permission.Support {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageInvalidUser)
		ctx.ReactWithCross()
		return
	}

	if err := logic.ClaimTicket(ctx.Worker, ticket, userId); err != nil {
		ctx.HandleError(err)
		return
	}

	mention := fmt.Sprintf("<@%d>", userId)
	ctx.SendEmbedNoDelete(utils.Green, "Ticket Claimed", translations.MessageClaimed, mention)
	ctx.ReactWithCheck()
}
