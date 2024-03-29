package admin

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/common/permission"
	w "github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest/ratelimit"
	"strconv"
)

type AdminRecacheCommand struct {
}

func (AdminRecacheCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "recache",
		Description:     i18n.HelpAdmin,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		Arguments: command.Arguments(
			command.NewOptionalArgument("guildid", "ID of the guild to recache", interaction.OptionTypeString, i18n.MessageInvalidArgument),
		),
	}
}

func (c AdminRecacheCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminRecacheCommand) Execute(ctx registry.CommandContext, providedGuildId *string) {
	var guildId uint64
	if providedGuildId != nil {
		var err error
		guildId, err = strconv.ParseUint(*providedGuildId, 10, 64)
		if err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		guildId = ctx.GuildId()
	}

	// purge cache
	ctx.Worker().Cache.DeleteGuild(guildId)
	ctx.Worker().Cache.DeleteGuildChannels(guildId)
	ctx.Worker().Cache.DeleteGuildRoles(guildId)

	// re-cache
	botId, isWhitelabel, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	var worker *w.Context
	if isWhitelabel {
		bot, err := dbclient.Client.Whitelabel.GetByBotId(botId)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if bot.BotId == 0 {
			ctx.HandleError(errors.New("bot not found"))
			return
		}

		worker = &w.Context{
			Token:        bot.Token,
			BotId:        bot.BotId,
			IsWhitelabel: true,
			ShardId:      0,
			Cache:        ctx.Worker().Cache,
			RateLimiter:  ratelimit.NewRateLimiter(ratelimit.NewRedisStore(redis.Client, fmt.Sprintf("ratelimiter:%d", bot.BotId)), 1),
		}
	} else {
		worker = ctx.Worker()
	}

	if _, err := worker.GetGuild(guildId); err != nil {
		ctx.HandleError(err)
		return
	}

	if _, err := worker.GetGuildChannels(guildId); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyPlainPermanent("done")
}
