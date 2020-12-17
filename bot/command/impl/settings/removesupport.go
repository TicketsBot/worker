package settings

import (
	permcache "github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type RemoveSupportCommand struct {
}

func (RemoveSupportCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "removesupport",
		Description:     translations.HelpRemoveSupport,
		Aliases:         []string{"removesuport"},
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewOptionalArgument("user", "User to remove the support representative permission from", interaction.OptionTypeUser, translations.MessageAddAdminNoMembers),
			command.NewOptionalArgument("role", "Role to remove the support representative permission from", interaction.OptionTypeRole, translations.MessageAddAdminNoMembers),
			command.NewOptionalArgumentMessageOnly("role_name", "Name of the role to remove the support representative permission from", interaction.OptionTypeString, translations.MessageAddAdminNoMembers),
		),
	}
}

func (c RemoveSupportCommand) GetExecutor() interface{} {
	return c.Execute
}

// TODO: Remove from existing tickets
func (RemoveSupportCommand) Execute(ctx command.CommandContext, userId *uint64, roleId *uint64, roleName *string) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!removesupport @User`\n`t!removesupport @Role`\n`t!removesupport role name`",
		Inline: false,
	}

	if userId == nil && roleId == nil && roleName == nil {
		ctx.ReplyWithFields(utils.Red, "Error", translations.MessageRemoveSupportNoMembers, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	// get guild object
	guild, err := ctx.Worker().GetGuild(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if userId != nil {
		if guild.OwnerId == *userId {
			ctx.Reply(utils.Red, "Error", translations.MessageOwnerMustBeAdmin)
			ctx.Reject()
			return
		}

		if ctx.UserId() == *userId {
			ctx.Reply(utils.Red, "Error", translations.MessageRemoveStaffSelf)
			ctx.Reject()
			return
		}

		if err := dbclient.Client.Permissions.RemoveSupport(ctx.GuildId(), *userId); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := permcache.SetCachedPermissionLevel(redis.Client, ctx.GuildId(), *userId, permcache.Everyone); err != nil {
			ctx.HandleError(err)
			return
		}
	}

	var roles []uint64
	if roleId != nil {
		roles = []uint64{*roleId}
	}

	if roleName != nil {
		guildRoles, err := ctx.Worker().GetGuildRoles(ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
			return
		}

		// Get role ID from name
		valid := false
		for _, role := range guildRoles {
			if strings.ToLower(role.Name) == strings.ToLower(*roleName) {
				valid = true
				roles = append(roles, role.Id)
				break
			}
		}

		// Verify a valid role was mentioned
		if !valid {
			ctx.ReplyWithFields(utils.Red, "Error", translations.MessageRemoveSupportNoMembers, utils.FieldsToSlice(usageEmbed))
			ctx.Reject()
			return
		}
	}

	// Add roles to DB
	for _, role := range roles {
		if err := dbclient.Client.RolePermissions.RemoveSupport(ctx.GuildId(), role); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := permcache.SetCachedPermissionLevel(redis.Client, ctx.GuildId(), role, permcache.Everyone); err != nil {
			ctx.HandleError(err)
			return
		}
	}

	ctx.Accept()
}
