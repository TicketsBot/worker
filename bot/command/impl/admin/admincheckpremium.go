package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type AdminCheckPremiumCommand struct {
}

func (AdminCheckPremiumCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "checkpremium",
		Description:     database.HelpAdminCheckPremium,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly: true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("guild_id", "ID of the guild to check premium status for", interaction.OptionTypeString, database.MessageInvalidArgument),
		),
	}
}

func (c AdminCheckPremiumCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminCheckPremiumCommand) Execute(ctx command.CommandContext, raw string) {
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

	tier := utils.PremiumClient.GetTierByGuild(guild, false)

	ctx.ReplyRaw(utils.Green, "Admin", fmt.Sprintf("`%s` has premium tier %d", guild.Name, tier))
	ctx.Accept()
}
