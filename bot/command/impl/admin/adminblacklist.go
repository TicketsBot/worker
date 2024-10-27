package admin

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
	"time"
)

type AdminBlacklistCommand struct {
}

func (AdminBlacklistCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "blacklist",
		Description:     i18n.HelpAdminBlacklist,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("guild_id", "ID of the guild to blacklist", interaction.OptionTypeString, i18n.MessageInvalidArgument),
			command.NewOptionalArgument("reason", "Reason for blacklisting the guild", interaction.OptionTypeString, i18n.MessageInvalidArgument),
		),
		Timeout: time.Second * 10,
	}
}

func (c AdminBlacklistCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminBlacklistCommand) Execute(ctx registry.CommandContext, raw string, reason *string) {
	guildId, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Error), "Invalid guild ID provided")
		return
	}

	if err := dbclient.Client.ServerBlacklist.Add(ctx, guildId, reason); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyPlainPermanent("🔨")

	// Check if whitelabel
	botId, ok, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(ctx, guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	var w *worker.Context
	if ok { // Whitelabel bot
		// Get bot
		bot, err := dbclient.Client.Whitelabel.GetByBotId(ctx, botId)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		w = &worker.Context{
			Token:        bot.Token,
			BotId:        bot.BotId,
			IsWhitelabel: true,
			Cache:        ctx.Worker().Cache,
			RateLimiter:  nil, // Use http-proxy ratelimit functionality
		}
	} else { // Public bot
		w = ctx.Worker()
	}

	if err := w.LeaveGuild(guildId); err != nil {
		ctx.HandleError(err)
		return
	}
}
