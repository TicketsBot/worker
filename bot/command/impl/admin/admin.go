package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"strings"
)

type AdminCommand struct {
}

func (AdminCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "admin",
		Description:     database.HelpAdmin,
		Aliases:         []string{"a"},
		PermissionLevel: permission.Everyone,
		Children: []registry.Command{
			AdminBlacklistCommand{},
			// AdminCheckPermsCommand{},
			AdminCheckPremiumCommand{},
			// AdminDebugCommand{},
			AdminForceCloseCommand{},
			AdminGenPremiumCommand{},
			AdminGetOwnerCommand{},
			AdminRecacheCommand{},
			// AdminPingCommand{},
			AdminSeedCommand{},
			AdminSetMessageCommand{},
			AdminUnblacklistCommand{},
			AdminUpdateSchemaCommand{},
			// AdminUsersCommand{},
		},
		Category:   command.Settings,
		HelperOnly: true,
		MessageOnly: true,
	}
}

func (c AdminCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminCommand) Execute(ctx registry.CommandContext) {
	msg := "Select a subcommand:\n"

	children := AdminCommand{}.Properties().Children
	for _, child := range children {
		if child.Properties().InteractionOnly {
			continue
		}

		description := i18n.GetMessageFromGuild(ctx.GuildId(), child.Properties().Description)
		msg += fmt.Sprintf("`%sadmin %s` - %s\n", utils.DEFAULT_PREFIX, child.Properties().Name, description)
	}

	msg = strings.TrimSuffix(msg, "\n")

	ctx.ReplyRaw(utils.Green, "Admin", msg)
}
