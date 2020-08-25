package admin

import (
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/rest/request"
	"strconv"
)

type AdminBlacklistCommand struct {
}

func (AdminBlacklistCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "blacklist",
		Description:     database.HelpAdminBlacklist,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:      true,
	}
}

func (AdminBlacklistCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbedRaw(utils.Red, "Error", "No guild ID provided")
		return
	}

	guildId, err := strconv.ParseUint(ctx.Args[0], 10, 64)
	if err != nil {
		ctx.SendEmbedRaw(utils.Red, "Error", "Invalid guild ID provided")
		return
	}

	if err := ctx.Worker.LeaveGuild(guildId); err != nil && !request.IsClientError(err) {
		ctx.HandleError(err)
		return
	}

	if err := dbclient.Client.ServerBlacklist.Add(guildId); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReactWithCheck()
}
