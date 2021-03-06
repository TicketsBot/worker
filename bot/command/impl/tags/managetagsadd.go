package tags

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
)

type ManageTagsAddCommand struct {
}

func (ManageTagsAddCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "add",
		Description:     translations.HelpTagAdd,
		Aliases:         []string{"new", "create"},
		PermissionLevel: permission.Support,
		Category:        command.Tags,
		InteractionOnly: true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("id", "Identifier for the tag", interaction.OptionTypeString, translations.MessageTagCreateInvalidArguments),
			command.NewRequiredArgument("content", "Tag contents to be sent when /tag is used", interaction.OptionTypeString, translations.MessageTagCreateInvalidArguments),
		),
	}
}

func (c ManageTagsAddCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ManageTagsAddCommand) Execute(ctx registry.CommandContext, tagId, content string) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!managetags add [TagID] [Tag contents]`",
		Inline: false,
	}

	// Length check
	if len(tagId) > 16 {
		ctx.Reject()
		ctx.ReplyWithFields(utils.Red, "Error", translations.MessageTagCreateTooLong, utils.FieldsToSlice(usageEmbed))
		return
	}

	// Verify a tag with the ID doesn't already exist
	// TODO: This causes a race condition, just try to insert and handle error
	var tagExists bool
	{
		tag, err := dbclient.Client.Tag.Get(ctx.GuildId(), tagId)
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.Reject()
			return
		}

		tagExists = tag != ""
	}

	if tagExists {
		ctx.ReplyWithFields(utils.Red, "Error", translations.MessageTagCreateAlreadyExists, utils.FieldsToSlice(usageEmbed), tagId, tagId)
		ctx.Reject()
		return
	}

	if err := dbclient.Client.Tag.Set(ctx.GuildId(), tagId, content); err == nil {
		ctx.Accept()
	} else {
		ctx.Reject()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}
