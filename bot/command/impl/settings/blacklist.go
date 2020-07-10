package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
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
	}
}

func (BlacklistCommand) Execute(ctx command.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!blacklist @User`",
		Inline: false,
	}

	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageBlacklistNoMembers, utils.FieldsToSlice(usageEmbed))
		ctx.ReactWithCross()
		return
	}

	user := ctx.Message.Mentions[0]
	user.Member.User = user.User

	if ctx.Author.Id == user.Id {
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageBlacklistSelf, utils.FieldsToSlice(usageEmbed))
		ctx.ReactWithCross()
		return
	}

	if permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker), user.Member, ctx.GuildId) > permission.Everyone {
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageBlacklistStaff, utils.FieldsToSlice(usageEmbed))
		ctx.ReactWithCross()
		return
	}

	isBlacklisted, err := dbclient.Client.Blacklist.IsBlacklisted(ctx.GuildId, user.Id)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		ctx.ReactWithCross()
		return
	}

	if isBlacklisted {
		if err := dbclient.Client.Blacklist.Remove(ctx.GuildId, user.Id); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.ReactWithCross()
			return
		}
	} else {
		if err := dbclient.Client.Blacklist.Add(ctx.GuildId, user.Id); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.ReactWithCross()
			return
		}
	}

	ctx.ReactWithCheck()
}
