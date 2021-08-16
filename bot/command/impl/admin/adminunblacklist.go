package admin

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type AdminUnblacklistCommand struct {
}

func (AdminUnblacklistCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "unblacklist",
		Description:     i18n.HelpAdminUnblacklist,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly:     true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("guild_id", "ID of the guild to unblacklist", interaction.OptionTypeString, i18n.MessageInvalidArgument),
		),
	}
}

func (c AdminUnblacklistCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminUnblacklistCommand) Execute(ctx registry.CommandContext, raw string) {
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
