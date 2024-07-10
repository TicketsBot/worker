package settings

import (
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/permission"
	"time"
)

type RemoveSupportCommand struct{}

func (RemoveSupportCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "removesupport",
		Description:     i18n.HelpRemoveSupport,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user_or_role", "User or role to remove the support representative permission from", interaction.OptionTypeMentionable, i18n.MessageRemoveSupportNoMembers),
		),
		DefaultEphemeral: true,
		Timeout:          time.Second * 5,
	}
}

func (c RemoveSupportCommand) GetExecutor() interface{} {
	return c.Execute
}

// TODO: Remove from existing tickets
func (c RemoveSupportCommand) Execute(ctx registry.CommandContext, id uint64) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`/removesupport @User`\n`/removesupport @Role`",
		Inline: false,
	}

	settings, err := ctx.Settings()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	mentionableType, valid := context.DetermineMentionableType(ctx, id)
	if !valid {
		ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageRemoveSupportNoMembers, utils.ToSlice(usageEmbed))
		return
	}

	if mentionableType == context.MentionableTypeUser {
		// get guild object
		guild, err := ctx.Worker().GetGuild(ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if guild.OwnerId == id {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageOwnerMustBeAdmin)
			return
		}

		if ctx.UserId() == id {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageRemoveStaffSelf)
			return
		}

		if err := dbclient.Client.Permissions.RemoveSupport(ctx, ctx.GuildId(), id); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx, ctx.GuildId(), id, permcache.Everyone); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := logic.RemoveOnCallRoles(ctx, ctx, id); err != nil {
			ctx.HandleError(err)
			return
		}
	} else if mentionableType == context.MentionableTypeRole {
		if err := dbclient.Client.RolePermissions.RemoveSupport(ctx, ctx.GuildId(), id); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx, ctx.GuildId(), id, permcache.Everyone); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := logic.RecreateOnCallRole(ctx, ctx, nil); err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		ctx.HandleError(fmt.Errorf("infallible"))
		return
	}

	ctx.Reply(customisation.Green, i18n.TitleRemoveSupport, i18n.MessageRemoveSupportSuccess)

	if settings.TicketNotificationChannel != nil {
		// Remove user / role from thread notification channel
		_ = ctx.Worker().EditChannelPermissions(*settings.TicketNotificationChannel, channel.PermissionOverwrite{
			Id:    id,
			Type:  mentionableType.OverwriteType(),
			Allow: 0,
			Deny:  permission.BuildPermissions(permission.ViewChannel),
		})
	}
}
