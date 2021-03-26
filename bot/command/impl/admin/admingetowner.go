package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type AdminGetOwnerCommand struct {
}

func (AdminGetOwnerCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "getowner",
		Description:     database.HelpAdminGetOwner,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly: true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("guild_id", "ID of the guild to get the owner of", interaction.OptionTypeString, database.MessageInvalidArgument),
		),
	}
}

func (c AdminGetOwnerCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminGetOwnerCommand) Execute(ctx registry.CommandContext, raw string) {
	guildId, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		ctx.ReplyRaw(utils.Red, "Error", "Invalid guild ID provided")
		return
	}

	guild, err := ctx.Worker().GetGuild(guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyRaw(utils.Green, "Admin", fmt.Sprintf("`%s` is owned by <@%d> (%d)", guild.Name, guild.OwnerId, guild.OwnerId))
	ctx.Accept()
}
