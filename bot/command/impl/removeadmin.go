package impl

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

type RemoveAdminCommand struct {
}

func (RemoveAdminCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "removeAdmin",
		Description:     "Revokes a user's or role's admin privileges",
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
	}
}

func (RemoveAdminCommand) Execute(ctx command.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!removeadmin @User`\n`t!removeadmin @Role`\n`t!removeadmin role name`",
		Inline: false,
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to revoke admin privileges from", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	// get guild object
	guild, err := ctx.Guild(); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	roles := make([]uint64, 0)
	if len(ctx.Message.Mentions) > 0 {
		for _, mention := range ctx.Message.Mentions {
			if guild.OwnerId == mention.Id {
				ctx.SendEmbed(utils.Red, "Error", "The guild owner must be an admin")
				continue
			}

			if ctx.Author.Id == mention.Id {
				ctx.SendEmbed(utils.Red, "Error", "You cannot revoke your own privileges")
				continue
			}

			go func() {
				if err := dbclient.Client.Permissions.RemoveAdmin(ctx.GuildId, mention.Id); err != nil {
					sentry.ErrorWithContext(err, ctx.ToErrorContext())
					ctx.ReactWithCross()
				}

				if err := permcache.SetCachedPermissionLevel(redis.Client, ctx.GuildId, mention.Id, permcache.Support); err != nil {
					ctx.HandleError(err)
					return
				}
			}()
		}
	} else if len(ctx.Message.MentionRoles) > 0 {
		for _, mention := range ctx.Message.MentionRoles {
			roles = append(roles, mention)
		}
	} else {
		roleName := strings.ToLower(strings.Join(ctx.Args, " "))

		// Get role ID from name
		valid := false
		for _, role := range guild.Roles {
			if strings.ToLower(role.Name) == roleName {
				roles = append(roles, role.Id)
				valid = true
				break
			}
		}

		// Verify a valid role was mentioned
		if !valid {
			ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to revoke admin privileges from", usageEmbed)
			ctx.ReactWithCross()
			return
		}
	}

	// Remove roles from DB
	for _, role := range roles {
		go func() {
			if err := dbclient.Client.RolePermissions.RemoveAdmin(ctx.GuildId, role); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
				ctx.ReactWithCross()
			}

			if err := permcache.SetCachedPermissionLevel(redis.Client, ctx.GuildId, role, permcache.Support); err != nil {
				ctx.HandleError(err)
				return
			}
		}()
	}

	ctx.ReactWithCheck()
}
