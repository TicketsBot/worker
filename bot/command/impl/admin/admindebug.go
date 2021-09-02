package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type AdminDebugCommand struct {
}

func (AdminDebugCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "debug",
		Description:     i18n.HelpAdminDebug,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly:     true,
	}
}

func (c AdminDebugCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminDebugCommand) Execute(ctx registry.CommandContext) {
	// Get ticket category
	categoryId, err := dbclient.Client.ChannelCategory.Get(ctx.GuildId())
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// get guild channels
	channels, err := ctx.Worker().GetGuildChannels(ctx.GuildId()); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	var categoryName string
	for _, channel := range channels {
		if channel.Id == categoryId { // Don't need to compare channel types
			categoryName = channel.Name
		}
	}

	if categoryName == "" {
		categoryName = "None"
	}

	// get guild object
	guild, err := ctx.Worker().GetGuild(ctx.GuildId()); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Get owner
	invalidOwner := false
	owner, err := ctx.Worker().GetGuildMember(ctx.GuildId(), guild.OwnerId); if err != nil {
		invalidOwner = true
	}

	var ownerFormatted string
	if invalidOwner {
		ownerFormatted = strconv.FormatUint(guild.OwnerId, 10)
	} else {
		ownerFormatted = fmt.Sprintf("%s#%s", owner.User.Username, utils.PadDiscriminator(owner.User.Discriminator))
	}

	embed := embed.NewEmbed().
		SetTitle("Admin").
		SetColor(int(constants.Green)).

		AddField("Shard", strconv.Itoa(ctx.Worker().ShardId), true).
		AddBlankField(false).

		AddField("Ticket Category", categoryName, true).
		AddField("Owner", ownerFormatted, true)

	ctx.ReplyWithEmbed(embed)
}
