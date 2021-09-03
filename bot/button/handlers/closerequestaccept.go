package handlers

import (
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
)

type CloseRequestAcceptHandler struct{}

func (h *CloseRequestAcceptHandler) Matcher() matcher.Matcher {
	return &matcher.SimpleMatcher{
		CustomId: "close_request_accept",
	}
}

func (h *CloseRequestAcceptHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed),
	}
}

func (h *CloseRequestAcceptHandler) Execute(ctx *context.ButtonContext) {
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.Id == 0 {
		ctx.Reply(constants.Red, "Error", i18n.MessageNotATicketChannel)
		return
	}

	if ctx.UserId() != ticket.UserId {
		ctx.Reply(constants.Red, "Error", i18n.MessageCloseRequestNoPermission)
		return
	}

	closeRequest, ok, err := dbclient.Client.CloseRequest.Get(ticket.GuildId, ticket.Id)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Infallible, unless malicious
	if !ok {
		return
	}

	// Create context for staff member - avoid users cant close issue
	newCtx := context.NewPanelContext(ctx.Worker(), ctx.GuildId(), ctx.ChannelId(), ticket.UserId, ctx.PremiumTier())
	logic.CloseTicket(&newCtx, closeRequest.Reason, true)
}
