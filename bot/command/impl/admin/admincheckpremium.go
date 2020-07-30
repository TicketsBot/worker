package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
	"strconv"
)

type AdminCheckPremiumCommand struct {
}

func (AdminCheckPremiumCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "getowner",
		Description:     database.HelpAdminCheckPremium,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
	}
}

func (AdminCheckPremiumCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbedRaw(utils.Red, "Error", "No guild ID provided")
		return
	}

	guildId, err := strconv.ParseUint(ctx.Args[0], 10, 64)
	if err != nil {
		ctx.SendEmbedRaw(utils.Red, "Error", "Invalid guild ID provided")
		return
	}

	guild, found := ctx.Worker.Cache.GetGuild(guildId, false)
	if !found {
		ctx.SendEmbedRaw(utils.Red, "Error", "Guild not found")
		return
	}

	tier := utils.PremiumClient.GetTierByGuild(guild, false)

	ctx.SendEmbedRaw(utils.Green, "Admin", fmt.Sprintf("`%s` has premium tier %d", guild.Name, tier))
	ctx.ReactWithCheck()
}
