package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
)

type AdminCheckPermsCommand struct {
}

func (AdminCheckPermsCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "checkperms",
		Description:     "Checks permissions for the bot on the channel",
		Aliases:         []string{"cp"},
		PermissionLevel: permission.Everyone,
		Parent:          AdminCommand{},
		Category:        command.Settings,
		HelperOnly:      true,
	}
}

func (AdminCheckPermsCommand) Execute(ctx command.CommandContext) {
	guild, err := ctx.Worker.GetGuild(ctx.GuildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendMessage(fmt.Sprintf("roles: %d", len(guild.Roles)))

	for _, role := range guild.Roles {
		ctx.SendMessage(fmt.Sprintf("role %s: %d", role.Name, role.Permissions))
	}
}
