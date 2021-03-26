package tags

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
)

type ManageTagsDeleteCommand struct {
}

func (ManageTagsDeleteCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "delete",
		Description:     translations.HelpTagDelete,
		Aliases:         []string{"del", "rm", "remove"},
		PermissionLevel: permission.Support,
		Category:        command.Tags,
		Arguments: command.Arguments(
			command.NewRequiredArgument("id", "ID of the tag to delete", interaction.OptionTypeString, translations.MessageTagDeleteInvalidArguments),
		),
	}
}

func (c ManageTagsDeleteCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ManageTagsDeleteCommand) Execute(ctx registry.CommandContext, tagId string) {
	/*usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!managetags delete [TagID]`",
		Inline: false,
	}*/

	// TODO: Causes a race condition, just try to delete
	var found bool
	{
		tag, err := dbclient.Client.Tag.Get(ctx.GuildId(), tagId)
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.Reject()
			return
		}

		found = tag != ""
	}

	if !found {
		ctx.Reject()
		ctx.Reply(utils.Red, "Error", translations.MessageTagDeleteDoesNotExist, tagId)
		return
	}

	if err := dbclient.Client.Tag.Delete(ctx.GuildId(), tagId); err == nil {
		ctx.Accept()
	} else {
		ctx.Reject()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}
