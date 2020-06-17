package admin

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"runtime"
)

type AdminGCCommand struct {
}

func (AdminGCCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "gc",
		Description:     "Forces a GC sweep",
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
	}
}

func (AdminGCCommand) Execute(ctx command.CommandContext) {
	runtime.GC()
	ctx.ReactWithCheck()
}
