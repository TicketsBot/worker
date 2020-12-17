package admin

import (
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type AdminUnblacklistCommand struct {
}

func (AdminUnblacklistCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "unblacklist",
		Description:     database.HelpAdminUnblacklist,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly:     true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("guild_id", "ID of the guild to unblacklist", interaction.OptionTypeString, database.MessageInvalidArgument),
		),
	}
}

func (c AdminUnblacklistCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminUnblacklistCommand) Execute(ctx command.CommandContext, raw string) {
	guildId, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		ctx.ReplyRaw(utils.Red, "Error", "Invalid guild ID provided")
		return
	}

	if err := dbclient.Client.ServerBlacklist.Delete(guildId); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Accept()
}
