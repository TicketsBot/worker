package impl

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/rest"
)

type UnclaimCommand struct {
}

func (UnclaimCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "unclaim",
		Description:     "Removes the claim on the current ticket",
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
	}
}

func (UnclaimCommand) Execute(ctx command.CommandContext) {
	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.SendEmbed(utils.Red, "Error", "This is not a ticket channel")
		ctx.ReactWithCross()
		return
	}

	// Get who claimed
	whoClaimed, err := dbclient.Client.TicketClaims.Get(ctx.GuildId, ticket.Id); if err != nil {
		ctx.HandleError(err)
		return
	}

	if whoClaimed == 0 {
		ctx.SendEmbed(utils.Red, "Error", "This ticket is not claimed")
		ctx.ReactWithCross()
		return
	}

	if ctx.UserPermissionLevel < permission.Admin && ctx.Author.Id != whoClaimed {
		ctx.SendEmbed(utils.Red, "Error", "Only admins and the user who claimed the ticket can unclaim the ticket")
		ctx.ReactWithCross()
		return
	}

	// Set to unclaimed in DB
	if err := dbclient.Client.TicketClaims.Delete(ctx.GuildId, ticket.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	// Update channel
	data := rest.ModifyChannelData{
		PermissionOverwrites: logic.CreateOverwrites(ctx.GuildId, ticket.UserId, ctx.Worker.BotId),
	}
	if _, err := ctx.Worker.ModifyChannel(ctx.ChannelId, data); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbed(utils.Green, "Ticket Unclaimed", "All support representatives can now respond to the ticket")
	ctx.ReactWithCheck()
}
