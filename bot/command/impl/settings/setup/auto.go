package setup

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	cmdcontext "github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	channel_permissions "github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"time"
)

const freePanelLimit = 3

type AutoSetupCommand struct {
}

func (AutoSetupCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "auto",
		Description:     i18n.HelpSetup,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Children:        nil,
		InteractionOnly: true,
		Timeout:         time.Second * 10,
	}
}

func (c AutoSetupCommand) GetExecutor() interface{} {
	return c.Execute
}

// TODO: Separate into diff functions
func (AutoSetupCommand) Execute(ctx registry.CommandContext) {
	interaction, ok := ctx.(*cmdcontext.SlashCommandContext)
	if !ok {
		return
	}

	var supportRoleId, adminRoleId uint64
	var failed bool
	var messageContent string

	// create roles
	switch role, err := ctx.Worker().CreateGuildRole(ctx.GuildId(), rest.GuildRoleData{Name: "Tickets Support"}); err {
	case nil: // an error occurred creating admin role
		supportRoleId = role.Id

		// set in db
		if err := dbclient.Client.RolePermissions.AddSupport(ctx, ctx.GuildId(), role.Id); err != nil {
			ctx.HandleError(err)
		}

		switch role, err := ctx.Worker().CreateGuildRole(ctx.GuildId(), rest.GuildRoleData{Name: "Tickets Admin"}); err {
		case nil:
			adminRoleId = role.Id

			// set in db
			if err := dbclient.Client.RolePermissions.AddAdmin(ctx, ctx.GuildId(), role.Id); err != nil {
				ctx.HandleError(err)
			}

			messageContent = fmt.Sprintf("✅ %s", i18n.GetMessageFromGuild(ctx.GuildId(), i18n.SetupAutoRolesSuccess))
		default:
			messageContent = fmt.Sprintf("❌ %s", i18n.GetMessageFromGuild(ctx.GuildId(), i18n.SetupAutoRolesFailure))
		}
	default: // an error occurred creating support role
		failed = true
		messageContent = fmt.Sprintf("❌ %s", i18n.GetMessageFromGuild(ctx.GuildId(), i18n.SetupAutoRolesFailure))
	}

	embed := embed.NewEmbed().
		SetTitle("Setup").
		SetColor(getColour(context.Background(), ctx.GuildId(), failed)). // TODO: Propagate context
		SetDescription(messageContent)

	ctx.ReplyWithEmbed(embed)

	// create transcripts channel
	switch transcriptChannel, err := ctx.Worker().CreateGuildChannel(ctx.GuildId(), getTranscriptChannelData(ctx.GuildId(), supportRoleId, adminRoleId)); err {
	case nil:
		messageContent += fmt.Sprintf("\n✅ %s", i18n.GetMessageFromGuild(ctx.GuildId(), i18n.SetupAutoTranscriptChannelSuccess, transcriptChannel.Id))

		if err := dbclient.Client.ArchiveChannel.Set(ctx, ctx.GuildId(), utils.Ptr(transcriptChannel.Id)); err != nil {
			ctx.HandleError(err)
		}
	default:
		failed = true
		messageContent += fmt.Sprintf("\n❌ %s", i18n.GetMessageFromGuild(ctx.GuildId(), i18n.SetupAutoTranscriptChannelFailure))
	}

	embed.SetDescription(messageContent)

	shouldEdit := true
	if err := edit(interaction, embed); err != nil {
		shouldEdit = false
	}

	// create category
	categoryData := rest.CreateChannelData{
		Name: "Tickets",
		Type: channel.ChannelTypeGuildCategory,
	}

	switch category, err := ctx.Worker().CreateGuildChannel(ctx.GuildId(), categoryData); err {
	case nil: // ok
		messageContent += fmt.Sprintf("\n✅ %s", i18n.GetMessageFromGuild(ctx.GuildId(), i18n.SetupAutoCategorySuccess))

		if err := dbclient.Client.ChannelCategory.Set(ctx, ctx.GuildId(), category.Id); err != nil {
			ctx.HandleError(err)
		}
	default: // error
		messageContent += fmt.Sprintf("\n❌ %s", i18n.GetMessageFromGuild(ctx.GuildId(), i18n.SetupAutoCategoryFailure))
	}

	messageContent += fmt.Sprintf("\n\n%s", i18n.GetMessageFromGuild(ctx.GuildId(), i18n.SetupAutoCompleted, ctx.GuildId(), adminRoleId, supportRoleId))
	messageContent += fmt.Sprintf("\n\n%s", i18n.GetMessageFromGuild(ctx.GuildId(), i18n.SetupAutoDocs))

	// update status
	if shouldEdit {
		embed.SetDescription(messageContent)

		if err := edit(interaction, embed); err != nil {
			shouldEdit = false
		}
	}
}

func edit(ctx *cmdcontext.SlashCommandContext, e *embed.Embed) error {
	data := rest.WebhookEditBody{
		Embeds: utils.Slice(e),
	}

	_, err := rest.EditOriginalInteractionResponse(context.Background(), ctx.Interaction.Token, ctx.Worker().RateLimiter, ctx.Worker().BotId, data)
	return err
}

func getColour(ctx context.Context, guildId uint64, failed bool) int {
	var colour customisation.Colour
	if failed {
		colour = customisation.Red
	} else {
		colour = customisation.Green
	}

	// ignore error, return default
	hex, _ := customisation.GetColour(ctx, guildId, colour)
	return hex
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
			Id:    guildId,
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
