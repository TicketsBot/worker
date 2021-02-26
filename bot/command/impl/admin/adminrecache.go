package admin

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/rxdn/gdl/objects/interaction"
)

type AdminRecacheCommand struct {
}

func (AdminRecacheCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "recache",
		Description:     translations.HelpAdmin,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly:     true,
		Arguments: command.Arguments(
			command.NewOptionalArgument("guildid", "ID of the guild to recache", interaction.OptionTypeString, translations.MessageInvalidArgument),
		),
	}
}

func (c AdminRecacheCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminRecacheCommand) Execute(ctx command.CommandContext, providedGuildId *uint64) {
	var guildId uint64
	if providedGuildId != nil {
		guildId = *providedGuildId
	} else {
		guildId = ctx.GuildId()
	}

	// purge cache
	ctx.Worker().Cache.DeleteGuild(guildId)
	ctx.Worker().Cache.DeleteGuildChannels(guildId)
	ctx.Worker().Cache.DeleteGuildRoles(guildId)

	// re-cache
	_, err := ctx.Worker().GetGuild(guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	_, err = ctx.Worker().GetGuildChannels(guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	_, err = ctx.Worker().GetGuildRoles(guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Accept()
}
