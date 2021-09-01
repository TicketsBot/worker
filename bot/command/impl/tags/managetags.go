package tags

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type ManageTagsCommand struct {
}

func (ManageTagsCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "managetags",
		Description:     i18n.HelpManageTags,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"managecannedresponse", "managecannedresponses", "editcannedresponse", "editcannedresponses", "ecr", "managetags", "mcr", "managetag", "mt"},
		PermissionLevel: permission.Support,
		Children: []registry.Command{
			ManageTagsAddCommand{},
			ManageTagsDeleteCommand{},
			ManageTagsListCommand{},
		},
		Category: command.Tags,
	}
}

func (c ManageTagsCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ManageTagsCommand) Execute(ctx registry.CommandContext) {
	msg := "Select a subcommand:\n"

	children := ManageTagsCommand{}.Properties().Children
	for _, child := range children {
		msg += fmt.Sprintf("`/managetags %s` - %s\n", child.Properties().Name, i18n.GetMessageFromGuild(ctx.GuildId(), child.Properties().Description))
	}

	msg = strings.TrimSuffix(msg, "\n")

	ctx.ReplyRaw(utils.Red, "Error", msg)
}
