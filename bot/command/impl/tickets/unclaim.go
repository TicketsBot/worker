package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/rest"
)

type UnclaimCommand struct {
}

func (UnclaimCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "unclaim",
		Description:     translations.HelpUnclaim,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
	}
}

func (c UnclaimCommand) GetExecutor() interface{} {
	return c.Execute
}

func (UnclaimCommand) Execute(ctx registry.CommandContext) {
	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId()); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.Reply(utils.Red, "Error", translations.MessageNotATicketChannel)
		ctx.Reject()
		return
	}

	// Get who claimed
	whoClaimed, err := dbclient.Client.TicketClaims.Get(ctx.GuildId(), ticket.Id); if err != nil {
		ctx.HandleError(err)
		return
	}

	if whoClaimed == 0 {
		ctx.Reply(utils.Red, "Error", translations.MessageNotClaimed)
		ctx.Reject()
		return
	}

	permissionLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permissionLevel < permission.Admin && ctx.UserId() != whoClaimed {
		ctx.Reply(utils.Red, "Error", translations.MessageOnlyClaimerCanUnclaim)
		ctx.Reject()
		return
	}

	// Set to unclaimed in DB
	if err := dbclient.Client.TicketClaims.Delete(ctx.GuildId(), ticket.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	// get panel
	var panel *database.Panel
	if ticket.PanelId != nil {
		var derefPanel database.Panel
		derefPanel, err = dbclient.Client.Panel.Get(*ticket.PanelId)

		if derefPanel.MessageId != 0 {
			panel = &derefPanel
		}
	}

	// Update channel
	data := rest.ModifyChannelData{
		PermissionOverwrites: logic.CreateOverwrites(ctx.Worker(), ctx.GuildId(), ticket.UserId, ctx.Worker().BotId, panel),
	}

	if _, err := ctx.Worker().ModifyChannel(ctx.ChannelId(), data); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(utils.Green, "Ticket Unclaimed", translations.MessageUnclaimed)
	ctx.Accept()
}
