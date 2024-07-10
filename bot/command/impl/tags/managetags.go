package tags

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
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
		Category:         command.Tags,
		DefaultEphemeral: true,
	}
}

func (c ManageTagsCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ManageTagsCommand) Execute(_ registry.CommandContext) {
	// Cannot call parent command
}
