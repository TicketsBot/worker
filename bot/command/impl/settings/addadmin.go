package settings

import (
	"context"
	permcache "github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
	"golang.org/x/sync/errgroup"
	"strings"
)

type AddAdminCommand struct {
	Registry registry.Registry
}

func (AddAdminCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "addadmin",
		Description:     translations.HelpAddAdmin,
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewOptionalArgument("user", "User to apply the administrator permission to", interaction.OptionTypeUser, translations.MessageAddAdminNoMembers),
			command.NewOptionalArgument("role", "Role to apply the administrator permission to", interaction.OptionTypeRole, translations.MessageAddAdminNoMembers),
			command.NewOptionalArgumentMessageOnly("role_name", "Name of the role to apply the administrator permission to", interaction.OptionTypeString, translations.MessageAddAdminNoMembers),
		),
	}
}

func (c AddAdminCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c AddAdminCommand) Execute(ctx registry.CommandContext, userId *uint64, roleId *uint64, roleName *string) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!addadmin @User`\n`t!addadmin @Role`\n`t!addadmin role name`",
		Inline: false,
	}

	if userId == nil && roleId == nil && roleName == nil {
		ctx.ReplyWithFields(utils.Red, "Error", translations.MessageAddAdminNoMembers, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	roles := make([]uint64, 0)

	if userId != nil {
		// Guild owner doesn't need to be added
		guild, err := ctx.Guild()
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if guild.OwnerId == *userId {
			ctx.Reply(utils.Red, "Error", translations.MessageOwnerIsAlreadyAdmin)
			return
		}

		if err := dbclient.Client.Permissions.AddAdmin(ctx.GuildId(), *userId); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), *userId, permcache.Admin); err != nil {
			ctx.HandleError(err)
			return
		}
	}

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
			if strings.ToLower(role.Name) == *roleName {
				valid = true
				roles = append(roles, role.Id)
				break
			}
		}

		// Verify a valid role was mentioned
		if !valid {
			ctx.ReplyWithFields(utils.Red, "Error", translations.MessageAddAdminNoMembers, utils.FieldsToSlice(usageEmbed))
			ctx.Reject()
			return
		}
	}

	// Add roles to DB
	group, _ := errgroup.WithContext(context.Background())
	for _, role := range roles {
		role := role

		group.Go(func() (err error) {
			if err = dbclient.Client.RolePermissions.AddAdmin(ctx.GuildId(), role); err != nil {
				return
			}

			if err = utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), role, permcache.Admin); err != nil {
				return
			}

			return
		})
	}

	if err := group.Wait(); err != nil {
		ctx.HandleError(err)
		return
	}

	//logic.UpdateCommandPermissions(ctx, c.Registry)

	openTickets, err := dbclient.Client.Tickets.GetGuildOpenTickets(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Update permissions for existing tickets
	for _, ticket := range openTickets {
		if ticket.ChannelId == nil {
			continue
		}

		ch, err := ctx.Worker().GetChannel(*ticket.ChannelId)
		if err != nil {
			// Check if the channel has been deleted
			if restError, ok := err.(request.RestError); ok && restError.StatusCode == 404 {
				if err := dbclient.Client.Tickets.CloseByChannel(*ticket.ChannelId); err != nil {
					ctx.HandleError(err)
					return
				}

				continue
			} else {
				ctx.HandleError(err)
				return
			}
		}

		overwrites := ch.PermissionOverwrites

		if userId != nil {
			overwrites = append(overwrites, channel.PermissionOverwrite{
				Id:    *userId,
				Type:  channel.PermissionTypeMember,
				Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
				Deny:  0,
			})
		}

		// If adding a role as an admin, apply overrides to role
		for _, role := range roles {
			overwrites = append(overwrites, channel.PermissionOverwrite{
				Id:    role,
				Type:  channel.PermissionTypeRole,
				Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
				Deny:  0,
			})
		}

		data := rest.ModifyChannelData{
			PermissionOverwrites: overwrites,
			Position:             ch.Position,
		}

		if _, err = ctx.Worker().ModifyChannel(*ticket.ChannelId, data); err != nil {
			ctx.HandleError(err)
			return
		}
	}

	ctx.Accept()
}
