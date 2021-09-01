package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type AdminCheckPermsCommand struct {
}

func (AdminCheckPermsCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "checkperms",
		Description:     i18n.HelpAdminCheckPerms,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"cp"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly:     true,
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
