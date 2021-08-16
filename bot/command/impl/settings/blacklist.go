package settings

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
)

type BlacklistCommand struct {
}

func (BlacklistCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "blacklist",
		Description:     i18n.HelpBlacklist,
		Aliases:         []string{"unblacklist"},
		PermissionLevel: permission.Support,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "User to blacklist or unblacklsit", interaction.OptionTypeUser, i18n.MessageBlacklistNoMembers),
		),
	}
}

func (c BlacklistCommand) GetExecutor() interface{} {
	return c.Execute
}

func (BlacklistCommand) Execute(ctx registry.CommandContext, userId uint64) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!blacklist @User`",
		Inline: false,
	}

	member, err := ctx.Worker().GetGuildMember(ctx.GuildId(), userId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ctx.UserId() == member.User.Id {
		ctx.ReplyWithFields(utils.Red, "Error", i18n.MessageBlacklistSelf, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	permLevel, err := permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker()), member, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permLevel > permission.Everyone {
		ctx.ReplyWithFields(utils.Red, "Error", i18n.MessageBlacklistStaff, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	isBlacklisted, err := dbclient.Client.Blacklist.IsBlacklisted(ctx.GuildId(), member.User.Id)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		ctx.Reject()
		return
	}

	if isBlacklisted {
		if err := dbclient.Client.Blacklist.Remove(ctx.GuildId(), member.User.Id); err != nil {
			ctx.HandleError(err)
			return
		}


		ctx.ReplyRaw(utils.Green, "Blacklist", fmt.Sprintf("<@%d> has been unblacklisted", member.User.Id))
	} else {
		if err := dbclient.Client.Blacklist.Add(ctx.GuildId(), member.User.Id); err != nil {
			ctx.HandleError(err)
			return
		}

		ctx.ReplyRaw(utils.Green, "Blacklist", fmt.Sprintf("<@%d> has been blacklisted", member.User.Id))
	}
}
