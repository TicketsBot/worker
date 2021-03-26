package admin

import (
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"runtime"
)

type AdminGCCommand struct {
}

func (AdminGCCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "gc",
		Description:     database.HelpAdminGC,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly: true,
	}
}

func (c AdminGCCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminGCCommand) Execute(ctx registry.CommandContext) {
	runtime.GC()
	ctx.Accept()
}
