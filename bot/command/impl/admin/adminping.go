package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
	"time"
)

type AdminPingCommand struct {
}

func (AdminPingCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "ping",
		Description:     database.HelpAdminPing,
		Aliases:         []string{"latency"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
	}
}

func (AdminPingCommand) Execute(ctx command.CommandContext) {
	latency := time.Now().Sub(ctx.Timestamp)
	ctx.SendEmbedRaw(utils.Green, "Admin", fmt.Sprintf("Shard %d latency: `%dms`", ctx.Worker.ShardId, latency))
}
