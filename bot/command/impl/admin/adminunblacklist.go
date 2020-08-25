package admin

import (
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
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
		AdminOnly:      true,
	}
}

func (AdminUnblacklistCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbedRaw(utils.Red, "Error", "No guild ID provided")
		return
	}

	guildId, err := strconv.ParseUint(ctx.Args[0], 10, 64)
	if err != nil {
		ctx.SendEmbedRaw(utils.Red, "Error", "Invalid guild ID provided")
		return
	}

	if err := dbclient.Client.ServerBlacklist.Delete(guildId); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReactWithCheck()
}
