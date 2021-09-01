package admin

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type AdminUpdateSchemaCommand struct {
}

func (AdminUpdateSchemaCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "updateschema",
		Description:     i18n.HelpAdminUpdateSchema,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly:     true,
	}
}

func (c AdminUpdateSchemaCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminUpdateSchemaCommand) Execute(ctx registry.CommandContext) {
	dbclient.Client.CreateTables(dbclient.Pool)
	ctx.Accept()
}
