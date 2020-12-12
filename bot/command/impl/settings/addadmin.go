package settings

import (
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"strings"
)

type AddAdminCommand struct {
}

func (AddAdminCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "addadmin",
		Description:     translations.HelpAddAdmin,
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewOptionalArgument("user", "User to apply the administrator permission to", interaction.OptionTypeUser, translations.MessageAddAdminNoMembers),
			command.NewOptionalArgument("role", "Role to apply the administrator permission to", interaction.OptionTypeRole, translations.MessageAddAdminNoMembers),
			command.NewOptionalArgumentMessageOnly("role_name", "Names of roles to apply the administrator permission to", interaction.OptionTypeString, translations.MessageAddAdminNoMembers),
		),
	}
}

func (c AddAdminCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AddAdminCommand) Execute(ctx command.CommandContext, user *member.Member, role *guild.Role, names *string) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!addadmin @User`\n`t!addadmin @Role`\n`t!addadmin role name`",
		Inline: false,
	}

	if user == nil && role == nil && names == nil {
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageAddAdminNoMembers, utils.FieldsToSlice(usageEmbed))
		ctx.ReactWithCross()
		return
	}

	roles := make([]uint64, 0)

	if user != nil {
		if err := dbclient.Client.Permissions.AddAdmin(ctx.GuildId, user.User.Id); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}

		if err := permcache.SetCachedPermissionLevel(redis.Client, ctx.GuildId, user.User.Id, permcache.Admin); err != nil {
			ctx.HandleError(err)
			return
		}
	}

	if role != nil {
		roles = []uint64{role.Id}
	}

	if names != nil {
		guildRoles, err := ctx.Worker.GetGuildRoles(ctx.GuildId)
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		roleName := strings.ToLower(strings.Join(ctx.Args, " "))

		// Get role ID from name
		valid := false
		for _, role := range guildRoles {
			if strings.ToLower(role.Name) == roleName {
				valid = true
				roles = append(roles, role.Id)
				break
			}
		}

		// Verify a valid role was mentioned
		if !valid {
			fmt.Println(1)
			ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageAddAdminNoMembers, utils.FieldsToSlice(usageEmbed))
			ctx.ReactWithCross()
			return
		}
	}

	// Add roles to DB
	for _, role := range roles {
		role := role

		go func() {
			if err := dbclient.Client.RolePermissions.AddAdmin(ctx.GuildId, role); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}

			if err := permcache.SetCachedPermissionLevel(redis.Client, ctx.GuildId, role, permcache.Admin); err != nil {
				ctx.HandleError(err)
				return
			}
		}()
	}

	openTickets, err := dbclient.Client.Tickets.GetGuildOpenTickets(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Update permissions for existing tickets
	for _, ticket := range openTickets {
		if ticket.ChannelId == nil {
			continue
		}

		ch, err := ctx.Worker.GetChannel(*ticket.ChannelId)
		if err != nil {
			continue
		}

		overwrites := ch.PermissionOverwrites

		if user != nil {
			// If adding individual admins, apply each override individually
			for _, mention := range ctx.Message.Mentions {
				overwrites = append(overwrites, channel.PermissionOverwrite{
					Id:    mention.Id,
					Type:  channel.PermissionTypeMember,
					Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
					Deny:  0,
				})
			}
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

		if _, err = ctx.Worker.ModifyChannel(*ticket.ChannelId, data); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
	}

	ctx.ReactWithCheck()
}
