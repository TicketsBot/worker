package impl

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/permission"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
)

type BlacklistCommand struct {
}

func (BlacklistCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "blacklist",
		Description:     "Toggles whether users are allowed to interact with the bot",
		Aliases:         []string{"unblacklist"},
		PermissionLevel: permcache.Support,
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
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user to toggle the blacklist state for", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	user := ctx.Message.Mentions[0]
	user.Member.User = user.User

	if ctx.Author.Id == user.Id {
		ctx.SendEmbed(utils.Red, "Error", "You cannot blacklist yourself")
		ctx.ReactWithCross()
		return
	}

	permissionLevelChan := make(chan permcache.PermissionLevel)
	go permission.GetPermissionLevel(ctx.Worker, user.Member, ctx.GuildId, permissionLevelChan)
	permissionLevel := <-permissionLevelChan

	if permissionLevel > permcache.Everyone {
		ctx.SendEmbed(utils.Red, "Error", "You cannot blacklist staff")
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
