package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
)

type ManageTagsDeleteCommand struct {
}

func (ManageTagsDeleteCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "delete",
		Description:     "Deletes a tag",
		Aliases:         []string{"del", "rm", "remove"},
		PermissionLevel: permission.Support,
		Parent:          ManageTagsCommand{},
		Category:        command.Tags,
	}
}

func (ManageTagsDeleteCommand) Execute(ctx command.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!managetags delete [TagID]`",
		Inline: false,
	}

	if len(ctx.Args) == 0 {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You must specify a tag ID to delete", usageEmbed)
		return
	}

	id := ctx.Args[0]

	var found bool
	{
		tag, err := dbclient.Client.Tag.Get(ctx.GuildId, id)
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.ReactWithCross()
			return
		}

		found = tag != ""
	}

	if !found {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("A tag with the ID `%s` could not be found", id))
		return
	}

	if err := dbclient.Client.Tag.Delete(ctx.GuildId, id); err == nil {
		ctx.ReactWithCheck()
	} else {
		ctx.ReactWithCross()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}
