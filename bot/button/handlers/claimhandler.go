package handlers

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
)

type ClaimHandler struct{}

func (h *ClaimHandler) Matches(customId string) bool {
	return customId == "claim"
}

func (h *ClaimHandler) Properties() registry.Properties {
	return registry.Properties{
		DMsAllowed: false,
	}
}

func (h *ClaimHandler) Execute(worker *worker.Context, data interaction.ButtonInteraction) {
	if data.Member == nil {
		return
	}

	errorCtx := errorcontext.WorkerErrorContext{
		Guild:   data.GuildId.Value,
		User:    data.Member.User.Id,
		Channel: data.ChannelId,
	}

	// TODO: Create a button context
	premiumTier, err := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, worker.Token, worker.RateLimiter)
	if err != nil {
		sentry.ErrorWithContext(err, errorCtx)
		return
	}

	// Get permission level
	permissionLevel, err := permission.GetPermissionLevel(utils.ToRetriever(worker), *data.Member, data.GuildId.Value)
	if err != nil {
		sentry.ErrorWithContext(err, errorCtx)
		return
	}

	if permissionLevel < permission.Support {
		utils.SendEmbed(worker, data.ChannelId, data.GuildId.Value, nil, utils.Red, "Error", i18n.MessageNoPermission, nil, 30, premiumTier > premium.None)
		//ctx.Reply(utils.Red, "Error", i18n.MessageNoPermission)
		return
	}

	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(data.ChannelId); if err != nil {
		sentry.ErrorWithContext(err, errorCtx)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		utils.SendEmbed(worker, data.ChannelId, data.GuildId.Value, nil, utils.Red, "Error", i18n.MessageNotATicketChannel, nil, 30, premiumTier > premium.None)
		//ctx.Reply(utils.Red, "Error", i18n.MessageNotATicketChannel)
		return
	}

	if err := logic.ClaimTicket(worker, ticket, data.Member.User.Id); err != nil {
		sentry.ErrorWithContext(err, errorCtx)
		//ctx.HandleError(err)
		return
	}

	utils.SendEmbed(worker, data.ChannelId, data.GuildId.Value, nil, utils.Green, "Ticket Claimed", i18n.MessageClaimed, nil, -1, premiumTier > premium.None, fmt.Sprintf("<@%d>", data.Member.User.Id))

}
