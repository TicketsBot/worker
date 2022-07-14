package settings

import (
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
)

type RemoveAdminCommand struct{}

func (RemoveAdminCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "removeadmin",
		Description:     i18n.HelpRemoveAdmin,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user_or_role", "User or role to remove the administrator permission from", interaction.OptionTypeMentionable, i18n.MessageRemoveAdminNoMembers),
		),
	}
}

func (c RemoveAdminCommand) GetExecutor() interface{} {
	return c.Execute
}

// TODO: Remove from existing tickets
func (c RemoveAdminCommand) Execute(ctx registry.CommandContext, id uint64) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`/removeadmin @User`\n`/removeadmin @Role`",
		Inline: false,
	}

	mentionableType, valid := context.DetermineMentionableType(ctx, id)
	if !valid {
		ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageRemoveAdminNoMembers, utils.ToSlice(usageEmbed))
		ctx.Reject()
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
			ctx.Reject()
			return
		}

		if ctx.UserId() == id {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageRemoveStaffSelf)
			ctx.Reject()
			return
		}

		if err := dbclient.Client.Permissions.RemoveAdmin(ctx.GuildId(), id); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), id, permcache.Support); err != nil {
			ctx.HandleError(err)
			return
		}
	} else if mentionableType == context.MentionableTypeRole {
		if err := dbclient.Client.RolePermissions.RemoveAdmin(ctx.GuildId(), id); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), id, permcache.Support); err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		ctx.HandleError(fmt.Errorf("infallible"))
		return
	}

	ctx.Accept()
	ctx.Reply(customisation.Green, i18n.TitleRemoveAdmin, i18n.MessageRemoveAdminSuccess)
}
