package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
	"time"
)

type AdminWhitelabelAssignGuildCommand struct {
}

func (AdminWhitelabelAssignGuildCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "whitelabel-assign-guild",
		Description:     i18n.HelpAdmin,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("bot_id", "ID of the bot to assign to the guild", interaction.OptionTypeString, i18n.MessageInvalidArgument),
			command.NewRequiredArgument("guild_id", "ID of the guild to assign the bot to", interaction.OptionTypeString, i18n.MessageInvalidArgument),
		),
		Timeout: time.Second * 10,
	}
}

func (c AdminWhitelabelAssignGuildCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminWhitelabelAssignGuildCommand) Execute(ctx registry.CommandContext, botIdRaw, guildIdRaw string) {
	botId, err := strconv.ParseUint(botIdRaw, 10, 64)
	if err != nil {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Error), "Invalid bot ID")
		return
	}

	guildId, err := strconv.ParseUint(guildIdRaw, 10, 64)
	if err != nil {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Error), "Invalid guild ID")
		return
	}

	bot, err := dbclient.Client.Whitelabel.GetByBotId(ctx, botId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if bot.BotId == 0 {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Error), "Whitelabel bot with provided ID not found")
		return
	}

	if err := dbclient.Client.WhitelabelGuilds.Add(ctx, botId, guildId); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyRaw(customisation.Green, ctx.GetMessage(i18n.Success), fmt.Sprintf("Assigned bot `%d` to guild `%d`", botId, guildId))
}
