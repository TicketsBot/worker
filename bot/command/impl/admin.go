package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
	"strings"
)

type AdminCommand struct {
}

func (AdminCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "admin",
		Description:     "Bot management",
		Aliases:         []string{"a"},
		PermissionLevel: permission.Everyone,
		Children: []command.Command{
			AdminCheckPermsCommand{},
			AdminDebugCommand{},
			AdminForceCloseCommand{},
			AdminGCCommand{},
			AdminGenPremiumCommand{},
			AdminGetOwnerCommand{},
			AdminPingCommand{},
			AdminSeedCommand{},
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
		msg += fmt.Sprintf("`%sadmin %s` - %s\n", utils.DEFAULT_PREFIX, child.Properties().Name, child.Properties().Description)
	}

	msg = strings.TrimSuffix(msg, "\n")

	ctx.SendEmbed(utils.Green, "Admin", msg)
}
