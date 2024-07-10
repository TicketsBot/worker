package handlers

import (
	"errors"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type JoinThreadHandler struct{}

func (h *JoinThreadHandler) Matcher() matcher.Matcher {
	return &matcher.FuncMatcher{
		Func: func(customId string) bool {
			return strings.HasPrefix(customId, "join_thread_")
		},
	}
}

func (h *JoinThreadHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.GuildAllowed),
		Timeout: time.Second * 5,
	}
}

var joinThreadPattern = regexp.MustCompile(`join_thread_(\d+)`)

func (h *JoinThreadHandler) Execute(ctx *context.ButtonContext) {
	groups := joinThreadPattern.FindStringSubmatch(ctx.InteractionData.CustomId)
	if len(groups) < 2 {
		return
	}

	// Errors are impossible
	ticketId, _ := strconv.Atoi(groups[1])

	// Get ticket
	ticket, err := dbclient.Client.Tickets.Get(ctx, ticketId, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !ticket.IsThread {
		ctx.HandleError(errors.New("Ticket is not a thread"))
		return
	}

	if !ticket.Open {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageJoinClosedTicket)

		// Try to delete message
		_ = ctx.Worker().DeleteMessage(ctx.ChannelId(), ctx.Interaction.Message.Id)

		return
	}

	if ticket.ChannelId == nil {
		ctx.HandleError(errors.New("Ticket channel not found"))
		return
	}

	// Check permission
	hasPermission, err := logic.HasPermissionForTicket(ctx, ctx.Worker(), ticket, ctx.UserId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !hasPermission {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageJoinThreadNoPermission)
		return
	}

	if _, err := ctx.Worker().GetThreadMember(*ticket.ChannelId, ctx.UserId()); err == nil {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageAlreadyJoinedThread, *ticket.ChannelId)
		return
	}

	// Join ticket
	if err := ctx.Worker().AddThreadMember(*ticket.ChannelId, ctx.UserId()); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(customisation.Green, i18n.Success, i18n.MessageJoinThreadSuccess, *ticket.ChannelId)
}
