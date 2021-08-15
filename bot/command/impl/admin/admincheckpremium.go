package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type AdminCheckPremiumCommand struct {
}

func (AdminCheckPremiumCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "checkpremium",
		Description:     i18n.HelpAdminCheckPremium,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly:     true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("guild_id", "ID of the guild to check premium status for", interaction.OptionTypeString, i18n.MessageInvalidArgument),
		),
	}
}

func (c AdminCheckPremiumCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminCheckPremiumCommand) Execute(ctx registry.CommandContext, raw string) {
	guildId, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		ctx.ReplyRaw(utils.Red, "Error", "Invalid guild ID provided")
		return
	}

	guild, found := ctx.Worker().Cache.GetGuild(guildId, false)
	if !found {
		ctx.ReplyRaw(utils.Red, "Error", "Guild not found")
		return
	}

	tier, src, err := utils.PremiumClient.GetTierByGuild(guild)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyRaw(utils.Green, "Admin", fmt.Sprintf("`%s` (owner %d) has premium tier %d (src %s)", guild.Name, guild.OwnerId, tier, src.String()))
	ctx.Accept()
}
