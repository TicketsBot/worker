package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
)

type AdminCheckPermsCommand struct {
}

func (AdminCheckPermsCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "checkperms",
		Description:     database.HelpAdminCheckPerms,
		Aliases:         []string{"cp"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly: true,
	}
}

func (c AdminCheckPermsCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminCheckPermsCommand) Execute(ctx registry.CommandContext) {
	guild, err := ctx.Worker().GetGuild(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyPlain(fmt.Sprintf("roles: %d", len(guild.Roles)))

	for _, role := range guild.Roles {
		ctx.ReplyPlain(fmt.Sprintf("role %s: %d", role.Name, role.Permissions))
	}
}
