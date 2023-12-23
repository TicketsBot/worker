package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type AdminCommand struct {
}

func (AdminCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "admin",
		Description:     i18n.HelpAdmin,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"a"},
		PermissionLevel: permission.Everyone,
		Children: []registry.Command{
			AdminBlacklistCommand{},
			AdminCheckPremiumCommand{},
			AdminGenPremiumCommand{},
			AdminGetOwnerCommand{},
			AdminRecacheCommand{},
			AdminUnblacklistCommand{},
			AdminWhitelabelAssignGuildCommand{},
			AdminWhitelabelDataCommand{},
		},
		Category:   command.Settings,
		HelperOnly: true,
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
		msg += fmt.Sprintf("`%sadmin %s` - %s\n", utils.DefaultPrefix, child.Properties().Name, description)
	}

	msg = strings.TrimSuffix(msg, "\n")
	ctx.ReplyRaw(customisation.Green, ctx.GetMessage(i18n.Admin), msg)
}
