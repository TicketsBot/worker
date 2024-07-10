package tags

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
	"time"
)

type ManageTagsListCommand struct {
}

func (ManageTagsListCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "list",
		Description:      i18n.HelpTagList,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permission.Support,
		Category:         command.Tags,
		DefaultEphemeral: true,
		Timeout:          time.Second * 3,
	}
}

func (c ManageTagsListCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ManageTagsListCommand) Execute(ctx registry.CommandContext) {
	ids, err := dbclient.Client.Tag.GetTagIds(ctx, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	var joined string
	for _, id := range ids {
		joined += fmt.Sprintf("• `%s`\n", id)
	}
	joined = strings.TrimSuffix(joined, "\n")

	ctx.Reply(customisation.Green, i18n.TitleTags, i18n.MessageTagList, joined, "/")
}
