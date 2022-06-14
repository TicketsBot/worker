package settings

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
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

type BlacklistCommand struct {
}

func (BlacklistCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "blacklist",
		Description:     i18n.HelpBlacklist,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"unblacklist"},
		PermissionLevel: permission.Support,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user_or_role", "User or role to blacklist or unblacklist", interaction.OptionTypeMentionable, i18n.MessageBlacklistNoMembers),
		),
	}
}

func (c BlacklistCommand) GetExecutor() interface{} {
	return c.Execute
}

func (BlacklistCommand) Execute(ctx registry.CommandContext, id uint64) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`/blacklist @User`\n`/blacklist @Role`",
		Inline: false,
	}

	mentionableType, valid := context.DetermineMentionableType(ctx, id)
	if !valid {
		ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageBlacklistNoMembers, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	if mentionableType == context.MentionableTypeUser {
		member, err := ctx.Worker().GetGuildMember(ctx.GuildId(), id)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if ctx.UserId() == id {
			ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageBlacklistSelf, utils.FieldsToSlice(usageEmbed))
			ctx.Reject()
			return
		}

		permLevel, err := permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker()), member, ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if permLevel > permission.Everyone {
			ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageBlacklistStaff, utils.FieldsToSlice(usageEmbed))
			ctx.Reject()
			return
		}

		isBlacklisted, err := dbclient.Client.Blacklist.IsBlacklisted(ctx.GuildId(), id)
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.Reject()
			return
		}

		if isBlacklisted {
			if err := dbclient.Client.Blacklist.Remove(ctx.GuildId(), id); err != nil {
				ctx.HandleError(err)
				return
			}

			ctx.Reply(customisation.Green, i18n.TitleBlacklist, i18n.MessageBlacklistRemove, id)
		} else {
			if err := dbclient.Client.Blacklist.Add(ctx.GuildId(), member.User.Id); err != nil {
				ctx.HandleError(err)
				return
			}

			ctx.Reply(customisation.Green, i18n.TitleBlacklist, i18n.MessageBlacklistAdd, member.User.Id)
		}
	} else if mentionableType == context.MentionableTypeRole {
		// Check if role is staff
		isSupport, err := dbclient.Client.RolePermissions.IsSupport(id)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if isSupport {
			ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageBlacklistStaff, utils.FieldsToSlice(usageEmbed)) // TODO: Does this need a new message?
			return
		}

		// Check if staff is part of any team
		isSupport, err = dbclient.Client.SupportTeamRoles.IsSupport(ctx.GuildId(), id)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if isSupport {
			ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageBlacklistStaff, utils.FieldsToSlice(usageEmbed)) // TODO: Does this need a new message?
			return
		}

		isBlacklisted, err := dbclient.Client.RoleBlacklist.IsBlacklisted(ctx.GuildId(), id)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if isBlacklisted {
			if err := dbclient.Client.RoleBlacklist.Remove(ctx.GuildId(), id); err != nil {
				ctx.HandleError(err)
				return
			}

			ctx.Reply(customisation.Green, i18n.TitleBlacklist, i18n.MessageBlacklistRemoveRole, id)
		} else {
			if err := dbclient.Client.RoleBlacklist.Add(ctx.GuildId(), id); err != nil {
				ctx.HandleError(err)
				return
			}

			ctx.Reply(customisation.Green, i18n.TitleBlacklist, i18n.MessageBlacklistAddRole, id)
		}
	} else {
		ctx.HandleError(fmt.Errorf("infallible"))
		return
	}
}
