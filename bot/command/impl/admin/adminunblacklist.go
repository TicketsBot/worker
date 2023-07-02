package admin

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type AdminUnblacklistCommand struct {
}

func (AdminUnblacklistCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "unblacklist",
		Description:     i18n.HelpAdminUnblacklist,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("guild_id", "ID of the guild to unblacklist", interaction.OptionTypeString, i18n.MessageInvalidArgument),
		),
	}
}

func (c AdminUnblacklistCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminUnblacklistCommand) Execute(ctx registry.CommandContext, raw string) {
	guildId, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Error), "Invalid guild ID provided")
		return
	}

	if err := dbclient.Client.ServerBlacklist.Delete(guildId); err != nil {
		ctx.HandleError(err)
		return
	}
}
