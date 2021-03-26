package admin

import (
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type AdminForceCloseCommand struct {
}

func (AdminForceCloseCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "forceclose",
		Description:     database.HelpAdminForceClose,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly: true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("guild_id", "ID of the guild of the ticket to close", interaction.OptionTypeString, database.MessageInvalidArgument),
			command.NewRequiredArgument("ticket_id", "ID of the ticket to close", interaction.OptionTypeInteger, database.MessageInvalidArgument),
		),
	}
}

func (c AdminForceCloseCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminForceCloseCommand) Execute(ctx registry.CommandContext, guildRaw string, ticketId int) {
	guildId, err := strconv.ParseUint(guildRaw, 10, 64)
	if err != nil {
		ctx.ReplyRaw(utils.Red, "Error", "Invalid guild ID provided")
		return
	}

	if err := dbclient.Client.Tickets.Close(ticketId, guildId); err != nil {
		ctx.HandleError(err)
	}

	ctx.Accept()
}
