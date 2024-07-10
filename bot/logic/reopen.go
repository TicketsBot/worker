package logic

import (
	"context"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
)

func ReopenTicket(ctx context.Context, cmd registry.CommandContext, ticketId int) {
	// Check ticket limit
	permLevel, err := cmd.UserPermissionLevel(ctx)
	if err != nil {
		cmd.HandleError(err)
		return
	}

	if permLevel == permission.Everyone {
		ticketLimit, err := dbclient.Client.TicketLimit.Get(ctx, cmd.GuildId())
		if err != nil {
			cmd.HandleError(err)
			return
		}

		// TODO: count()
		openTickets, err := dbclient.Client.Tickets.GetOpenByUser(ctx, cmd.GuildId(), cmd.UserId())
		if err != nil {
			cmd.HandleError(err)
			return
		}

		if len(openTickets) >= int(ticketLimit) {
			cmd.Reply(customisation.Green, i18n.Error, i18n.MessageTicketLimitReached)
			return
		}
	}

	ticket, err := dbclient.Client.Tickets.Get(ctx, ticketId, cmd.GuildId())
	if err != nil {
		cmd.HandleError(err)
		return
	}

	if ticket.Id == 0 || ticket.GuildId != cmd.GuildId() {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageReopenTicketNotFound)
		return
	}

	// Ensure user has permissino to reopen the ticket
	hasPermission, err := HasPermissionForTicket(ctx, cmd.Worker(), ticket, cmd.UserId())
	if err != nil {
		cmd.HandleError(err)
		return
	}

	if !hasPermission {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageReopenNoPermission)
		return
	}

	// Ticket must be closed already
	if ticket.Open {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageReopenAlreadyOpen)
		return
	}

	// Only allow reopening threads
	if !ticket.IsThread {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageReopenNotThread)
		return
	}

	// Ensure channel still exists
	if ticket.ChannelId == nil {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageReopenThreadDeleted)
		return
	}

	ch, err := cmd.Worker().GetChannel(*ticket.ChannelId)
	if err != nil {
		if err, ok := err.(request.RestError); ok && err.StatusCode == 404 {
			cmd.Reply(customisation.Red, i18n.Error, i18n.MessageReopenThreadDeleted)
			return
		}
	}

	if ch.Id == 0 {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageReopenThreadDeleted)
		return
	}

	data := rest.ModifyChannelData{
		ThreadMetadataModifyData: &rest.ThreadMetadataModifyData{
			Archived: utils.Ptr(false),
			Locked:   utils.Ptr(false),
		},
	}

	if _, err := cmd.Worker().ModifyChannel(*ticket.ChannelId, data); err != nil {
		if err, ok := err.(request.RestError); ok && err.StatusCode == 404 {
			cmd.Reply(customisation.Red, i18n.Error, i18n.MessageReopenThreadDeleted)
			return
		}

		cmd.HandleError(err)
		return
	}

	cmd.Reply(customisation.Green, i18n.Success, i18n.MessageReopenSuccess, ticket.Id, *ticket.ChannelId)

	embedData := utils.BuildEmbed(cmd, customisation.Green, i18n.TitleReopened, i18n.MessageReopenedTicket, nil, cmd.UserId())
	if _, err := cmd.Worker().CreateMessageEmbed(*ticket.ChannelId, embedData); err != nil {
		cmd.HandleError(err)
		return
	}
}
