package tags

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type TagCommand struct {
}

func (c TagCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "tag",
		Description:     i18n.HelpTag,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"canned", "cannedresponse", "cr", "tags", "tag", "snippet", "c"},
		PermissionLevel: permission.Support,
		Category:        command.Tags,
		Arguments: command.Arguments(
			command.NewRequiredAutocompleteableArgument("id", "The ID of the tag to be sent to the channel", interaction.OptionTypeString, i18n.MessageTagInvalidArguments, c.AutoCompleteHandler),
		),
	}
}

func (c TagCommand) GetExecutor() interface{} {
	return c.Execute
}

func (TagCommand) Execute(ctx registry.CommandContext, tagId string) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!tag [TagID]`",
		Inline: false,
	}

	content, err := dbclient.Client.Tag.Get(ctx.GuildId(), tagId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		ctx.Reject()
		return
	}

	if content == "" {
		ctx.ReplyWithFields(constants.Red, i18n.Error, i18n.MessageTagInvalidTag, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	if ticket.UserId != 0 {
		mention := fmt.Sprintf("<@%d>", ticket.UserId)
		content = strings.Replace(content, "%user%", mention, -1)
	}

	// TODO: Delete message if message context
	//_ = ctx.Worker().DeleteMessage(ctx.ChannelId(), ctx.Id)

	ctx.ReplyPlainPermanent(content)
}

func (TagCommand) AutoCompleteHandler(data interaction.ApplicationCommandAutoCompleteInteraction, value string) []interaction.ApplicationCommandOptionChoice {
	tagIds, err := dbclient.Client.Tag.GetStartingWith(data.GuildId.Value, value, 25)
	if err != nil {
		sentry.Error(err) // TODO: Error context
		return nil
	}

	choices := make([]interaction.ApplicationCommandOptionChoice, len(tagIds))
	for i, tagId := range tagIds {
		choices[i] = utils.StringChoice(tagId)
	}

	return choices
}
