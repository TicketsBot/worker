package handlers

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"regexp"
	"strconv"
	"strings"
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
		Flags: registry.SumFlags(registry.GuildAllowed),
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
	ticket, err := dbclient.Client.Tickets.Get(ticketId, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !ticket.IsThread {
		ctx.HandleError(errors.New("Ticket is not a thread"))
		return
	}

	if !ticket.Open {
		ctx.ReplyPlain("Ticket is closed")

		// Try to delete message
		_ = ctx.Worker().DeleteMessage(ctx.ChannelId(), ctx.Interaction.Message.Id)

		return
	}

	if ticket.ChannelId == nil {
		ctx.HandleError(errors.New("Ticket channel not found"))
		return
	}

	// Check permission
	hasPermission, err := logic.HasPermissionForTicket(ctx.Worker(), ticket, ctx.UserId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !hasPermission {
		ctx.ReplyPlain("You do not have permission to join this ticket")
		return
	}

	// TODO: Check if already joined

	// Join ticket
	if err := ctx.Worker().AddThreadMember(*ticket.ChannelId, ctx.UserId()); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyPlain(fmt.Sprintf("<#%d>", *ticket.ChannelId))
}
