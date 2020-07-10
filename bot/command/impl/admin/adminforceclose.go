package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"strconv"
)

type AdminForceCloseCommand struct {
}

func (AdminForceCloseCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "forceclose",
		Description:     "Sets the state of the provided tickets to closed",
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
	}
}

func (AdminForceCloseCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) < 2 {
		ctx.SendEmbedRaw(utils.Red, "Error", "No guild ID provided")
		return
	}

	guildId, err := strconv.ParseUint(ctx.Args[0], 10, 64)
	if err != nil {
		ctx.SendEmbedRaw(utils.Red, "Error", "Invalid guild ID provided")
		return
	}

	for i := 1; i < len(ctx.Args); i++ {
		id, err := strconv.Atoi(ctx.Args[i])
		if err != nil {
			ctx.SendEmbedRaw(utils.Red, "Error", fmt.Sprintf("Invalid ticket ID provided: `%s`", ctx.Args[i]))
			continue
		}

		if err := dbclient.Client.Tickets.Close(id, guildId); err != nil {
			ctx.HandleError(err)
		}
	}

	ctx.ReactWithCheck()
}
