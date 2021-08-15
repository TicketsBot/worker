package settings

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type RemoveSupportCommand struct {
	Registry registry.Registry
}

func (RemoveSupportCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "removesupport",
		Description:     i18n.HelpRemoveSupport,
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewOptionalArgument("user", "User to remove the support representative permission from", interaction.OptionTypeUser, i18n.MessageAddAdminNoMembers),
			command.NewOptionalArgument("role", "Role to remove the support representative permission from", interaction.OptionTypeRole, i18n.MessageAddAdminNoMembers),
			command.NewOptionalArgumentMessageOnly("role_name", "Name of the role to remove the support representative permission from", interaction.OptionTypeString, i18n.MessageAddAdminNoMembers),
		),
	}
}

func (c RemoveSupportCommand) GetExecutor() interface{} {
	return c.Execute
}

// TODO: Remove from existing tickets
func (c RemoveSupportCommand) Execute(ctx registry.CommandContext, userId *uint64, roleId *uint64, roleName *string) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!removesupport @User`\n`t!removesupport @Role`\n`t!removesupport role name`",
		Inline: false,
	}

	if userId == nil && roleId == nil && roleName == nil {
		ctx.ReplyWithFields(utils.Red, "Error", i18n.MessageRemoveSupportNoMembers, utils.FieldsToSlice(usageEmbed))
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
			ctx.Reply(utils.Red, "Error", i18n.MessageOwnerMustBeAdmin)
			ctx.Reject()
			return
		}

		if ctx.UserId() == *userId {
			ctx.Reply(utils.Red, "Error", i18n.MessageRemoveStaffSelf)
			ctx.Reject()
			return
		}

		if err := dbclient.Client.Permissions.RemoveSupport(ctx.GuildId(), *userId); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), *userId, permcache.Everyone); err != nil {
			ctx.HandleError(err)
			return
		}
	}

	ctx.ReplyRaw(utils.Green, "Remove Support", "Support Representative removed successfully")

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
			ctx.ReplyWithFields(utils.Red, "Error", i18n.MessageRemoveSupportNoMembers, utils.FieldsToSlice(usageEmbed))
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

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), role, permcache.Everyone); err != nil {
			ctx.HandleError(err)
			return
		}
	}

	//logic.UpdateCommandPermissions(ctx, c.Registry)

	ctx.Accept()
}
