package settings

import (
	"context"
	permcache "github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"golang.org/x/sync/errgroup"
	"strings"
)

type RemoveAdminCommand struct {
}

func (RemoveAdminCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "removeadmin",
		Description:     translations.HelpRemoveAdmin,
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewOptionalArgument("user", "User to remove the administrator permission from", interaction.OptionTypeUser, translations.MessageAddAdminNoMembers),
			command.NewOptionalArgument("role", "Role to remove the administrator permission from", interaction.OptionTypeRole, translations.MessageAddAdminNoMembers),
			command.NewOptionalArgumentMessageOnly("role_name", "Name of the role to remove the administrator permission from", interaction.OptionTypeString, translations.MessageAddAdminNoMembers),
		),
	}
}

func (c RemoveAdminCommand) GetExecutor() interface{} {
	return c.Execute
}

// TODO: Remove from existing tickets
func (RemoveAdminCommand) Execute(ctx command.CommandContext, userId *uint64, roleId *uint64, roleName *string) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!removeadmin @User`\n`t!removeadmin @Role`\n`t!removeadmin role name`",
		Inline: false,
	}

	if userId == nil && roleId == nil && roleName == nil {
		ctx.ReplyWithFields(utils.Red, "Error", translations.MessageRemoveAdminNoMembers, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	// get guild object
	guild, err := ctx.Worker().GetGuild(ctx.GuildId()); if err != nil {
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

		if err := dbclient.Client.Permissions.RemoveAdmin(ctx.GuildId(), *userId); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), *userId, permcache.Support); err != nil {
			ctx.HandleError(err)
			return
		}
	}

	var roles []uint64
	if roleId != nil {
		roles = []uint64{*roleId}
	}

	if roleName != nil {
		// Get role ID from name
		valid := false
		for _, role := range guild.Roles {
			if strings.ToLower(role.Name) == strings.ToLower(*roleName) {
				roles = append(roles, role.Id)
				valid = true
				break
			}
		}

		// Verify a valid role was mentioned
		if !valid {
			ctx.ReplyWithFields(utils.Red, "Error", translations.MessageRemoveAdminNoMembers, utils.FieldsToSlice(usageEmbed))
			ctx.Reject()
			return
		}
	}

	// Remove roles from DB
	group, _ := errgroup.WithContext(context.Background())
	for _, role := range roles {
		role := role

		group.Go(func() error {
			if err := dbclient.Client.RolePermissions.RemoveAdmin(ctx.GuildId(), role); err != nil {
				return err
			}

			if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), role, permcache.Support); err != nil {
				return err
			}

			return nil
		})
	}

	switch group.Wait() {
	case nil:
		ctx.Accept()
	case err:
		ctx.HandleError(err)
	}
}
