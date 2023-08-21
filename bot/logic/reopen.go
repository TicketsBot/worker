package logic

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
)

func ReopenTicket(ctx registry.CommandContext, ticketId int) {
	// Check ticket limit
	permLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permLevel == permission.Everyone {
		ticketLimit, err := dbclient.Client.TicketLimit.Get(ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
			return
		}

		// TODO: count()
		openTickets, err := dbclient.Client.Tickets.GetOpenByUser(ctx.GuildId(), ctx.UserId())
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if len(openTickets) >= int(ticketLimit) {
			ctx.Reply(customisation.Green, i18n.Error, i18n.MessageTicketLimitReached)
			return
		}
	}

	ticket, err := dbclient.Client.Tickets.Get(ticketId, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.Id == 0 || ticket.GuildId != ctx.GuildId() {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageReopenTicketNotFound)
		return
	}

	// Ensure user has permissino to reopen the ticket
	hasPermission, err := HasPermissionForTicket(ctx.Worker(), ticket, ctx.UserId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !hasPermission {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageReopenNoPermission)
		return
	}

	// Ticket must be closed already
	if ticket.Open {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageReopenAlreadyOpen)
		return
	}

	// Only allow reopening threads
	if !ticket.IsThread {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageReopenNotThread)
		return
	}

	// Ensure channel still exists
	if ticket.ChannelId == nil {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageReopenThreadDeleted)
		return
	}

	ch, err := ctx.Worker().GetChannel(*ticket.ChannelId)
	if err != nil {
		if err, ok := err.(request.RestError); ok && err.StatusCode == 404 {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageReopenThreadDeleted)
			return
		}
	}

	if ch.Id == 0 {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageReopenThreadDeleted)
		return
	}

	data := rest.ModifyChannelData{
		ThreadMetadataModifyData: &rest.ThreadMetadataModifyData{
			Archived: utils.Ptr(false),
			Locked:   utils.Ptr(false),
		},
	}

	if _, err := ctx.Worker().ModifyChannel(*ticket.ChannelId, data); err != nil {
		if err, ok := err.(request.RestError); ok && err.StatusCode == 404 {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageReopenThreadDeleted)
			return
		}

		ctx.HandleError(err)
		return
	}

	ctx.Reply(customisation.Green, i18n.Success, i18n.MessageReopenSuccess, ticket.Id, *ticket.ChannelId)

	embedData := utils.BuildEmbed(ctx, customisation.Green, i18n.TitleReopened, i18n.MessageReopenedTicket, nil, ctx.UserId())
	if _, err := ctx.Worker().CreateMessageEmbed(*ticket.ChannelId, embedData); err != nil {
		ctx.HandleError(err)
		return
	}
}
