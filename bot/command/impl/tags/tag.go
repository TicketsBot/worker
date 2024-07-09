package tags

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
)

type TagCommand struct {
}

func (c TagCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "tag",
		Description:     i18n.HelpTag,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"canned", "cannedresponse", "cr", "tags", "tag", "snippet", "c"},
		PermissionLevel: permission.Everyone,
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
		Value:  "`/tag [TagID]`",
		Inline: false,
	}

	tag, ok, err := dbclient.Client.Tag.Get(ctx.GuildId(), tagId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !ok {
		ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageTagInvalidTag, utils.ToSlice(usageEmbed))
		return
	}

	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx.Worker().Context, ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Count user as a participant so that Tickets Answered stat includes tickets where only /tag was used
	if ticket.GuildId != 0 {
		go func() {
			if err := dbclient.Client.Participants.Set(ctx.Worker().Context, ctx.GuildId(), ticket.Id, ctx.UserId()); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}
		}()
	}

	content := utils.ValueOrZero(tag.Content)
	if ticket.Id != 0 {
		content = logic.DoPlaceholderSubstitutions(content, ctx.Worker(), ticket, nil)
	}

	var embeds []*embed.Embed
	if tag.Embed != nil {
		embeds = []*embed.Embed{
			logic.BuildCustomEmbed(ctx.Worker(), ticket, *tag.Embed.CustomEmbed, tag.Embed.Fields, false, nil),
		}
	}

	data := command.MessageResponse{
		Content: content,
		Embeds:  embeds,
	}

	if _, err := ctx.ReplyWith(data); err != nil {
		ctx.HandleError(err)
		return
	}
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
