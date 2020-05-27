package permission

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/permission"
)

func GetPermissionLevel(worker *worker.Context, member member.Member, guildId uint64, ch chan permcache.PermissionLevel) {
	// Check user ID in cache
	if cached, found := permcache.GetPermissionLevel(redis.Client, guildId, member.User.Id); found {
		ch <- cached
		return
	}

	// Check if the user is a bot admin user
	if utils.IsBotAdmin(member.User.Id) {
		ch <- permcache.Admin
		return
	}

	// Check if user is guild owner
	guild, err := worker.GetGuild(guildId)
	if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild: guildId,
			User:  member.User.Id,
		})
	}

	if err == nil {
		if member.User.Id == guild.OwnerId {
			go permcache.SetPermissionLevel(redis.Client, guildId, member.User.Id, permcache.Admin)
			ch <- permcache.Admin
			return
		}
	}

	// Check user perms for admin
	adminUser, err := dbclient.Client.Permissions.IsAdmin(guildId, member.User.Id); if err != nil {
		sentry.Error(err)
	}

	if adminUser {
		go permcache.SetPermissionLevel(redis.Client, guildId, member.User.Id, permcache.Admin)
		ch <- permcache.Admin
		return
	}

	// Check roles from DB
	adminRoles, err := dbclient.Client.RolePermissions.GetAdminRoles(guildId); if err != nil {
		sentry.Error(err)
	}

	for _, adminRoleId := range adminRoles {
		if member.HasRole(adminRoleId) {
			go permcache.SetPermissionLevel(redis.Client, guildId, member.User.Id, permcache.Admin)
			ch <- permcache.Admin
			return
		}
	}

	// Check if user has Administrator permission
	hasAdminPermission := HasPermissions(worker, guildId, member.User.Id, permission.Administrator)
	if hasAdminPermission {
		go permcache.SetPermissionLevel(redis.Client, guildId, member.User.Id, permcache.Admin)
		ch <- permcache.Admin
		return
	}

	// Check user perms for support
	isSupport, err := dbclient.Client.Permissions.IsSupport(guildId, member.User.Id); if err != nil {
		sentry.Error(err)
	}

	if isSupport {
		go permcache.SetPermissionLevel(redis.Client, guildId, member.User.Id, permcache.Support)
		ch <- permcache.Support
		return
	}

	// Check DB for support roles
	supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(guildId); if err != nil {
		sentry.Error(err)
	}

	for _, supportRoleId := range supportRoles {
		if member.HasRole(supportRoleId) {
			go permcache.SetPermissionLevel(redis.Client, guildId, member.User.Id, permcache.Support)
			ch <- permcache.Support
			return
		}
	}

	go permcache.SetPermissionLevel(redis.Client, guildId, member.User.Id, permcache.Everyone)
	ch <- permcache.Everyone
}
