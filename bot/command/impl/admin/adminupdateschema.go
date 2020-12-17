package admin

import (
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
)

type AdminUpdateSchemaCommand struct {
}

func (AdminUpdateSchemaCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "updateschema",
		Description:     database.HelpAdminUpdateSchema,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly: true,
	}
}

func (c AdminUpdateSchemaCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminUpdateSchemaCommand) Execute(ctx command.CommandContext) {
	dbclient.Client.CreateTables(dbclient.Pool)
	ctx.Accept()
}
