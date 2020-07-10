package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
	"strconv"
)

type AdminGetOwnerCommand struct {
}

func (AdminGetOwnerCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "getowner",
		Description:     "Gets the owner of a server",
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
	}
}

func (AdminGetOwnerCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbedRaw(utils.Red, "Error", "No guild ID provided")
		return
	}

	guildId, err := strconv.ParseUint(ctx.Args[0], 10, 64)
	if err != nil {
		ctx.SendEmbedRaw(utils.Red, "Error", "Invalid guild ID provided")
		return
	}

	guild, err := ctx.Worker.GetGuild(guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbedRaw(utils.Green, "Admin", fmt.Sprintf("`%s` is owned by <@%d> (%d)", guild.Name, guild.OwnerId, guild.OwnerId))
	ctx.ReactWithCheck()
}
