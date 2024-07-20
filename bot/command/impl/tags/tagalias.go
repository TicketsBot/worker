package tags

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type TagAliasCommand struct {
	tag database.Tag
}

func NewTagAliasCommand(tag database.Tag) TagAliasCommand {
	return TagAliasCommand{
		tag: tag,
	}
}

func (c TagAliasCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            c.tag.Id,
		Description:     i18n.HelpTag,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Tags,
		Timeout:         time.Second * 5,
	}
}

func (c TagAliasCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c TagAliasCommand) Execute(ctx registry.CommandContext) {
	if ctx.PremiumTier() < premium.Premium {
		ctx.Reply(customisation.Red, i18n.TitlePremiumOnly, i18n.MessageTagAliasRequiresPremium)
		return
	}

	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx, ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Count user as a participant so that Tickets Answered stat includes tickets where only /tag was used
	if ticket.GuildId != 0 {
		go func() {
			if err := dbclient.Client.Participants.Set(ctx, ctx.GuildId(), ticket.Id, ctx.UserId()); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}
		}()
	}

	content := utils.ValueOrZero(c.tag.Content)
	if ticket.Id != 0 {
		content = logic.DoPlaceholderSubstitutions(ctx, content, ctx.Worker(), ticket, nil)
	}

	var embeds []*embed.Embed
	if c.tag.Embed != nil {
		embeds = []*embed.Embed{
			logic.BuildCustomEmbed(ctx, ctx.Worker(), ticket, *c.tag.Embed.CustomEmbed, c.tag.Embed.Fields, false, nil),
		}
	}

	var allowedMentions message.AllowedMention
	if ticket.Id != 0 {
		allowedMentions = message.AllowedMention{
			Users: []uint64{ticket.UserId},
		}
	}

	data := command.MessageResponse{
		Content:         content,
		Embeds:          embeds,
		AllowedMentions: allowedMentions,
	}

	if _, err := ctx.ReplyWith(data); err != nil {
		ctx.HandleError(err)
		return
	}
}
