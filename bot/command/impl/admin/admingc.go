package admin

import (
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"runtime"
)

type AdminGCCommand struct {
}

func (AdminGCCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "gc",
		Description:     database.HelpAdminGC,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
	}
}

func (c AdminGCCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminGCCommand) Execute(ctx command.CommandContext) {
	runtime.GC()
	ctx.Accept()
}
