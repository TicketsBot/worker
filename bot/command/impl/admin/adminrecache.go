package admin

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type AdminRecacheCommand struct {
}

func (AdminRecacheCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "recache",
		Description:     i18n.HelpAdmin,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly:     true,
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
