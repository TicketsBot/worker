package settings

import (
	"fmt"
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

type BlacklistCommand struct {
}

func (BlacklistCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "blacklist",
		Description:     translations.HelpBlacklist,
		Aliases:         []string{"unblacklist"},
		PermissionLevel: permission.Support,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "User to blacklist or unblacklsit", interaction.OptionTypeUser, translations.MessageBlacklistNoMembers),
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
		ctx.ReplyWithFields(utils.Red, "Error", translations.MessageBlacklistSelf, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	permLevel, err := permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker()), member, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permLevel > permission.Everyone {
		ctx.ReplyWithFields(utils.Red, "Error", translations.MessageBlacklistStaff, utils.FieldsToSlice(usageEmbed))
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
