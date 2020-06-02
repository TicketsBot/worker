package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
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
		Description:     "Lists all tags",
		PermissionLevel: permission.Support,
		Parent:          ManageTagsCommand{},
		Category:        command.Tags,
	}
}

func (ManageTagsListCommand) Execute(ctx command.CommandContext) {
	ids, err := dbclient.Client.Tag.GetTagIds(ctx.GuildId)
	if err != nil {
		ctx.ReactWithCross()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	var joined string
	for _, id := range ids {
		joined += fmt.Sprintf("• `%s`\n", id)
	}
	joined = strings.TrimSuffix(joined, "\n")

	ctx.SendEmbed(utils.Green, "Tags", fmt.Sprintf("IDs for all tags:\n%s\nTo view the contents of a tag, run `%stag <ID>`", joined, utils.DEFAULT_PREFIX))
}
