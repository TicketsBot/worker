package handlers

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
)

type ClaimHandler struct{}

func (h *ClaimHandler) Matcher() matcher.Matcher {
	return &matcher.SimpleMatcher{
		CustomId: "claim",
	}
}

func (h *ClaimHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed),
	}
}

func (h *ClaimHandler) Execute(ctx *context.ButtonContext) {
	// Get permission level
	permissionLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permissionLevel < permission.Support {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageClaimNoPermission)
		return
	}

	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageNotATicketChannel)
		return
	}

	if err := logic.ClaimTicket(ctx.Worker(), ticket, ctx.UserId()); err != nil {
		ctx.HandleError(err)
		return
	}

	// TODO: Can we use ReplyWith?
	utils.SendEmbed(ctx.Worker(), ctx.ChannelId(), ctx.GuildId(), nil, constants.Green, "Ticket Claimed", i18n.MessageClaimed, nil, -1, ctx.PremiumTier() > premium.None, fmt.Sprintf("<@%d>", ctx.UserId()))
}
