package tags

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
	"strings"
)

type ManageTagsCommand struct {
}

func (ManageTagsCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "managetags",
		Description:     translations.HelpManageTags,
		Aliases:         []string{"managecannedresponse", "managecannedresponses", "editcannedresponse", "editcannedresponses", "ecr", "managetags", "mcr", "managetag", "mt"},
		PermissionLevel: permission.Support,
		Children: []command.Command{
			ManageTagsAddCommand{},
			ManageTagsDeleteCommand{},
			ManageTagsListCommand{},
		},
		Category:        command.Tags,
	}
}

func (ManageTagsCommand) Execute(ctx command.CommandContext) {
	msg := "Select a subcommand:\n"

	children := ManageTagsCommand{}.Properties().Children
	for _, child := range children {
		msg += fmt.Sprintf("`%smt %s` - %s\n", utils.DEFAULT_PREFIX, child.Properties().Name, child.Properties().Description)
	}

	msg = strings.TrimSuffix(msg, "\n")

	ctx.SendEmbedRaw(utils.Red, "Error", msg)
}
