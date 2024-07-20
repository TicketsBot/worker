package tags

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type ManageTagsAddCommand struct {
}

func (ManageTagsAddCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "add",
		Description:     i18n.HelpTagAdd,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"new", "create"},
		PermissionLevel: permission.Support,
		Category:        command.Tags,
		InteractionOnly: true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("id", "Identifier for the tag", interaction.OptionTypeString, i18n.MessageTagCreateInvalidArguments),
			command.NewRequiredArgument("content", "Tag contents to be sent when /tag is used", interaction.OptionTypeString, i18n.MessageTagCreateInvalidArguments),
		),
		DefaultEphemeral: true,
		Timeout:          time.Second * 3,
	}
}

func (c ManageTagsAddCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ManageTagsAddCommand) Execute(ctx registry.CommandContext, tagId, content string) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`/managetags add [TagID] [Tag Contents]`",
		Inline: false,
	}

	// Limit of 200 tags
	count, err := dbclient.Client.Tag.GetTagCount(ctx, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if count >= 200 {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageTagCreateLimit, 200)
		return
	}

	// Length check
	if len(tagId) > 16 {
		ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageTagCreateTooLong, utils.ToSlice(usageEmbed))
		return
	}

	// Verify a tag with the ID doesn't already exist
	exists, err := dbclient.Client.Tag.Exists(ctx, ctx.GuildId(), tagId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if exists {
		ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageTagCreateAlreadyExists, utils.ToSlice(usageEmbed), tagId, tagId)
		return
	}

	tag := database.Tag{
		Id:                   tagId,
		GuildId:              ctx.GuildId(),
		Content:              &content,
		Embed:                nil,
		ApplicationCommandId: nil,
	}

	if err := dbclient.Client.Tag.Set(ctx, tag); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(customisation.Green, i18n.MessageTag, i18n.MessageTagCreateSuccess, tagId)
}
