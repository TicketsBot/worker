package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"strings"
)

type AdminCommand struct {
}

func (AdminCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "admin",
		Description:     database.HelpAdmin,
		Aliases:         []string{"a"},
		PermissionLevel: permission.Everyone,
		Children: []command.Command{
			AdminBlacklistCommand{},
			AdminCheckPermsCommand{},
			AdminCheckPremiumCommand{},
			AdminDebugCommand{},
			AdminForceCloseCommand{},
			AdminGCCommand{},
			AdminGenPremiumCommand{},
			AdminGetOwnerCommand{},
			AdminPingCommand{},
			AdminSeedCommand{},
			AdminSetMessageCommand{},
			AdminUnblacklistCommand{},
			AdminUpdateSchemaCommand{},
			AdminUsersCommand{},
		},
		Category:   command.Settings,
		HelperOnly: true,
	}
}

func (AdminCommand) Execute(ctx command.CommandContext) {
	msg := "Select a subcommand:\n"

	children := AdminCommand{}.Properties().Children
	for _, child := range children {
		description := i18n.GetMessageFromGuild(ctx.GuildId, child.Properties().Description)
		msg += fmt.Sprintf("`%sadmin %s` - %s\n", utils.DEFAULT_PREFIX, child.Properties().Name, description)
	}

	msg = strings.TrimSuffix(msg, "\n")

	ctx.SendEmbedRaw(utils.Green, "Admin", msg)
}
