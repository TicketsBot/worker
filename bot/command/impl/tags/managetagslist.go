package tags

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"strings"
)

type ManageTagsListCommand struct {
}

func (ManageTagsListCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "list",
		Description:     translations.HelpTagList,
		PermissionLevel: permission.Support,
		Category:        command.Tags,
	}
}

func (ManageTagsListCommand) Execute(ctx command.CommandContext) {
	ids, err := dbclient.Client.Tag.GetTagIds(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	var joined string
	for _, id := range ids {
		joined += fmt.Sprintf("• `%s`\n", id)
	}
	joined = strings.TrimSuffix(joined, "\n")

	ctx.Reply(utils.Green, "Tags", translations.MessageTagList, joined, utils.DEFAULT_PREFIX)
}
