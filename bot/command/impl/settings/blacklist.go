package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
)

type BlacklistCommand struct {
}

func (BlacklistCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "blacklist",
		Description:     translations.HelpBlacklist,
		Aliases:         []string{"unblacklist"},
		PermissionLevel: permission.Support,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "User to blacklist or unblacklsit", interaction.OptionTypeUser),
		),
	}
}

func (c BlacklistCommand) GetExecutor() interface{} {
	return c.Execute
}

func (BlacklistCommand) Execute(ctx command.CommandContext, member member.Member) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!blacklist @User`",
		Inline: false,
	}

	if ctx.Author.Id == member.User.Id {
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageBlacklistSelf, utils.FieldsToSlice(usageEmbed))
		ctx.ReactWithCross()
		return
	}

	if permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker), member, ctx.GuildId) > permission.Everyone {
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageBlacklistStaff, utils.FieldsToSlice(usageEmbed))
		ctx.ReactWithCross()
		return
	}

	isBlacklisted, err := dbclient.Client.Blacklist.IsBlacklisted(ctx.GuildId, member.User.Id)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		ctx.ReactWithCross()
		return
	}

	if isBlacklisted {
		if err := dbclient.Client.Blacklist.Remove(ctx.GuildId, member.User.Id); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.ReactWithCross()
			return
		}
	} else {
		if err := dbclient.Client.Blacklist.Add(ctx.GuildId, member.User.Id); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.ReactWithCross()
			return
		}
	}

	ctx.ReactWithCheck()
}
