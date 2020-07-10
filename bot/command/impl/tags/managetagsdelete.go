package tags

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
)

type ManageTagsDeleteCommand struct {
}

func (ManageTagsDeleteCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "delete",
		Description:     translations.HelpTagDelete,
		Aliases:         []string{"del", "rm", "remove"},
		PermissionLevel: permission.Support,
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
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageTagDeleteInvalidArguments, []embed.EmbedField{usageEmbed})
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
		ctx.SendEmbed(utils.Red, "Error", translations.MessageTagDeleteDoesNotExist, id)
		return
	}

	if err := dbclient.Client.Tag.Delete(ctx.GuildId, id); err == nil {
		ctx.ReactWithCheck()
	} else {
		ctx.ReactWithCross()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}
