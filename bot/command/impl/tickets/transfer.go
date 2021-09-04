package tickets

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type TransferCommand struct {
}

func (TransferCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "transfer",
		Description:     i18n.HelpTransfer,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "Support representative to transfer the ticket to", interaction.OptionTypeUser, i18n.MessageInvalidUser),
		),
	}
}

func (c TransferCommand) GetExecutor() interface{} {
	return c.Execute
}

func (TransferCommand) Execute(ctx registry.CommandContext, userId uint64) {
	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageNotATicketChannel)
		ctx.Reject()
		return
	}

	member, err := ctx.Worker().GetGuildMember(ctx.GuildId(), userId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	permissionLevel, err := permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker()), member, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permissionLevel < permission.Support {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageInvalidUser)
		ctx.Reject()
		return
	}

	if err := logic.ClaimTicket(ctx.Worker(), ticket, userId); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyPermanent(constants.Green, i18n.TitleClaim, i18n.MessageClaimed, fmt.Sprintf("<@%d>", userId))
}
