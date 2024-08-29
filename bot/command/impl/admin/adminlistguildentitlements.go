package admin

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/common/permission"
	w "github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
	"time"
)

type AdminListGuildEntitlementsCommand struct {
}

func (AdminListGuildEntitlementsCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "list-guild-entitlements",
		Description:     i18n.HelpAdmin,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("guild_id", "Guild ID to fetch entitlements for", interaction.OptionTypeString, i18n.MessageInvalidArgument),
		),
		Timeout: time.Second * 15,
	}
}

func (c AdminListGuildEntitlementsCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminListGuildEntitlementsCommand) Execute(ctx registry.CommandContext, guildIdRaw string) {
	guildId, err := strconv.ParseUint(guildIdRaw, 10, 64)
	if err != nil {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Error), "Invalid guild ID provided")
		return
	}

	botId, isWhitelabel, err := dbclient.Client.WhitelabelGuilds.GetBotByGuild(ctx, guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Get guild
	var worker *w.Context
	if isWhitelabel {
		bot, err := dbclient.Client.Whitelabel.GetByBotId(ctx, botId)
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
			RateLimiter:  nil, // Use http-proxy ratelimit functionality
		}
	} else {
		worker = ctx.Worker()
	}

	guild, err := worker.GetGuild(guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// List entitlements that have expired in the past 30 days
	entitlements, err := dbclient.Client.Entitlements.ListGuildSubscriptions(ctx, guildId, guild.OwnerId, time.Hour*24*30)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	embed := embed.NewEmbed().
		SetTitle("Entitlements").
		SetColor(ctx.GetColour(customisation.Blue))

	if len(entitlements) == 0 {
		embed.SetDescription("No entitlements found")
	}

	for i, entitlement := range entitlements {
		if i >= 25 {
			embed.SetDescription("Too many entitlements to display")
			break
		}

		value := fmt.Sprintf(
			"**Tier:** %s\n**Source:** %s\n**Expires:** <t:%d>\n**SKU ID:** %s\n**SKU Priority:** %d",
			entitlement.Tier,
			entitlement.Source,
			entitlement.ExpiresAt.Unix(),
			entitlement.SkuId.String(),
			entitlement.SkuPriority,
		)

		embed.AddField(entitlement.SkuLabel, value, false)
	}

	ctx.ReplyWithEmbed(embed)
}
