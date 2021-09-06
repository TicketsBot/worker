package handlers

import (
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
)

type CloseRequestDenyHandler struct{}

func (h *CloseRequestDenyHandler) Matcher() matcher.Matcher {
	return &matcher.SimpleMatcher{
		CustomId: "close_request_deny",
	}
}

func (h *CloseRequestDenyHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed),
	}
}

func (h *CloseRequestDenyHandler) Execute(ctx *context.ButtonContext) {
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.Id == 0 {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageNotATicketChannel)
		return
	}

	if ctx.UserId() != ticket.UserId {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageCloseRequestNoPermission)
		return
	}

	if err := dbclient.Client.CloseRequest.Delete(ctx.GuildId(), ticket.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Edit(command.MessageResponse{
		Embeds: utils.Embeds(utils.BuildEmbed(ctx, constants.Red, i18n.TitleCloseRequest, i18n.MessageCloseRequestDenied, nil, ctx.UserId())),
	})
}
