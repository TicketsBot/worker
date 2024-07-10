package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
)

type UnclaimCommand struct {
}

func (UnclaimCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "unclaim",
		Description:     i18n.HelpUnclaim,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
		Timeout:         constants.TimeoutOpenTicket,
	}
}

func (c UnclaimCommand) GetExecutor() interface{} {
	return c.Execute
}

func (UnclaimCommand) Execute(ctx *context.SlashCommandContext) {
	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx, ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNotATicketChannel)
		return
	}

	// Check if thread
	if ticket.IsThread {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageClaimThread)
		return
	}

	// Get who claimed
	whoClaimed, err := dbclient.Client.TicketClaims.Get(ctx, ctx.GuildId(), ticket.Id)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if whoClaimed == 0 {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNotClaimed)
		return
	}

	permissionLevel, err := ctx.UserPermissionLevel(ctx)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permissionLevel < permission.Admin && ctx.UserId() != whoClaimed {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageOnlyClaimerCanUnclaim)
		return
	}

	// Set to unclaimed in DB
	if err := dbclient.Client.TicketClaims.Delete(ctx, ctx.GuildId(), ticket.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	// get panel
	var panel *database.Panel
	if ticket.PanelId != nil {
		var derefPanel database.Panel
		derefPanel, err = dbclient.Client.Panel.GetById(ctx, *ticket.PanelId)

		if derefPanel.PanelId != 0 {
			panel = &derefPanel
		}
	}

	overwrites, err := logic.CreateOverwrites(ctx.Context, ctx, ticket.UserId, panel)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Update channel
	data := rest.ModifyChannelData{
		PermissionOverwrites: overwrites,
	}

	if _, err := ctx.Worker().ModifyChannel(ctx.ChannelId(), data); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyPermanent(customisation.Green, i18n.TitleUnclaimed, i18n.MessageUnclaimed)
}
