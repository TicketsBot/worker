package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
	"time"
)

type AdminPingCommand struct {
}

func (AdminPingCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "ping",
		Description:     "Measures WS latency to Discord",
		Aliases:         []string{"latency"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
	}
}

func (AdminPingCommand) Execute(ctx command.CommandContext) {
	latency := time.Now().Sub(ctx.Timestamp)
	ctx.SendEmbed(utils.Green, "Admin", fmt.Sprintf("Shard %d latency: `%dms`", ctx.Worker.ShardId, latency))
}
