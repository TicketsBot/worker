package impl

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
)

type AdminUpdateSchemaCommand struct {
}

func (AdminUpdateSchemaCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "updateschema",
		Description:     "Updates the database schema",
		PermissionLevel: permission.Everyone,
		Parent:          AdminCommand{},
		Category:        command.Settings,
		AdminOnly:       true,
	}
}

func (AdminUpdateSchemaCommand) Execute(ctx command.CommandContext) {
	dbclient.Client.CreateTables(dbclient.Pool)
	ctx.ReactWithCheck()
}
