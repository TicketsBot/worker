package setup

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	channel_permissions "github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
)

type AutoSetupCommand struct {
}

func (AutoSetupCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "auto",
		Description:     translations.HelpSetup,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Children:        nil,
	}
}

func (c AutoSetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AutoSetupCommand) Execute(ctx registry.CommandContext) {
	var supportRoleId, adminRoleId uint64
	var messageId uint64
	var failed bool
	var messageContent string

	// create roles
	switch role, err := ctx.Worker().CreateGuildRole(ctx.GuildId(), supportRoleData); err {
	case nil: // an error occurred creating admin role
		supportRoleId = role.Id

		// set in db
		if err := dbclient.Client.RolePermissions.AddSupport(ctx.GuildId(), role.Id); err != nil {
			ctx.HandleError(err)
		}

		switch role, err := ctx.Worker().CreateGuildRole(ctx.GuildId(), adminRoleData); err {
		case nil:
			adminRoleId = role.Id

			// set in db
			if err := dbclient.Client.RolePermissions.AddAdmin(ctx.GuildId(), role.Id); err != nil {
				ctx.HandleError(err)
			}

			messageContent = fmt.Sprintf("✅ %s", i18n.GetMessageFromGuild(ctx.GuildId(), translations.SetupAutoRolesSuccess))
		default:
			messageContent = fmt.Sprintf("❌ %s", i18n.GetMessageFromGuild(ctx.GuildId(), translations.SetupAutoRolesFailure))
		}
	default: // an error occurred creating support role
		failed = true
		messageContent = fmt.Sprintf("❌ %s", i18n.GetMessageFromGuild(ctx.GuildId(), translations.SetupAutoRolesFailure))
	}

	embed := embed.NewEmbed().
		SetTitle("Setup").
		SetColor(getColour(failed)).
		SetDescription(messageContent)

	msg, _ := ctx.Worker().CreateMessageEmbed(ctx.ChannelId(), embed)
	messageId = msg.Id

	// create transcripts channel
	switch transcriptChannel, err := ctx.Worker().CreateGuildChannel(ctx.GuildId(), getTranscriptChannelData(ctx.GuildId(), supportRoleId, adminRoleId)); err {
	case nil:
		messageContent += fmt.Sprintf("\n✅ %s", i18n.GetMessageFromGuild(ctx.GuildId(), translations.SetupAutoTranscriptChannelSuccess, transcriptChannel.Id))

		if err := dbclient.Client.ArchiveChannel.Set(ctx.GuildId(), transcriptChannel.Id); err != nil {
			ctx.HandleError(err)
		}
	default:
		failed = true
		messageContent += fmt.Sprintf("\n❌ %s", i18n.GetMessageFromGuild(ctx.GuildId(), translations.SetupAutoTranscriptChannelFailure))
	}

	// update status
	if messageId != 0 {
		embed.SetDescription(messageContent)

		data := rest.EditMessageData{
			Embed: embed,
		}

		_, _ = ctx.Worker().EditMessage(ctx.ChannelId(), messageId, data)
	}

	// create category
	categoryData := rest.CreateChannelData{
		Name: "Tickets",
		Type: channel.ChannelTypeGuildCategory,
	}

	switch category, err := ctx.Worker().CreateGuildChannel(ctx.GuildId(), categoryData); err {
	case nil: // ok
		messageContent += fmt.Sprintf("\n✅ %s", i18n.GetMessageFromGuild(ctx.GuildId(), translations.SetupAutoCategorySuccess))

		if err := dbclient.Client.ChannelCategory.Set(ctx.GuildId(), category.Id); err != nil {
			ctx.HandleError(err)
		}
	default: // error
		messageContent += fmt.Sprintf("\n❌ %s", i18n.GetMessageFromGuild(ctx.GuildId(), translations.SetupAutoCategoryFailure))
	}

	{
		messageContent += fmt.Sprintf("\n%s", i18n.GetMessageFromGuild(ctx.GuildId(), translations.SetupAutoCompleted, ctx.GuildId(), adminRoleId, supportRoleId))
	}

	// update status
	if messageId != 0 {
		embed.SetDescription(messageContent)

		data := rest.EditMessageData{
			Embed: embed,
		}

		_, _ = ctx.Worker().EditMessage(ctx.ChannelId(), messageId, data)
	}
}

var (
	adminRoleData = rest.GuildRoleData{
		Name: "Tickets Admin",
	}
	supportRoleData = rest.GuildRoleData{
		Name: "Tickets Support",
	}
)

func getColour(failed bool) int {
	if failed {
		return int(utils.Red)
	} else {
		return int(utils.Green)
	}
}

func getTranscriptChannelData(guildId, supportRoleId, adminRoleId uint64) rest.CreateChannelData {
	allow := channel_permissions.BuildPermissions(
		channel_permissions.ViewChannel,
		channel_permissions.SendMessages,
		channel_permissions.EmbedLinks,
		channel_permissions.AttachFiles,
		channel_permissions.ReadMessageHistory,
	)

	overwrites := []channel.PermissionOverwrite{
		{ // deny everyone else access to channel
			Id:
			guildId,
			Type:  channel.PermissionTypeRole,
			Allow: 0,
			Deny:  allow,
		},
	}

	if supportRoleId != 0 {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    supportRoleId,
			Type:  channel.PermissionTypeRole,
			Allow: allow,
			Deny:  0,
		})
	}

	if adminRoleId != 0 {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    adminRoleId,
			Type:  channel.PermissionTypeRole,
			Allow: allow,
			Deny:  0,
		})
	}

	return rest.CreateChannelData{
		Name:                 "transcripts",
		Type:                 channel.ChannelTypeGuildText,
		PermissionOverwrites: overwrites,
	}
}
