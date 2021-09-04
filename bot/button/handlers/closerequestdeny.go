package handlers

import (
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/rest"
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

	messageId, err := dbclient.Client.CloseRequest.Delete(ctx.GuildId(), ticket.Id)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if messageId == 0 {
		return
	}

	data := rest.EditMessageData{
		Embed: utils.BuildEmbed(ctx, constants.Red, "Close Request", i18n.MessageCloseRequestDenied, nil, ctx.UserId()),
	}

	if _, err := ctx.Worker().EditMessage(ctx.ChannelId(), messageId, data); err != nil {
		ctx.HandleError(err)
		return
	}
}
