package permissionwrapper

import (
	"errors"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/permission"
)

func HasPermissionsChannel(ctx *worker.Context, guildId, userId, channelId uint64, permissions ...permission.Permission) bool {
	sum, err := getEffectivePermissionsChannel(ctx, guildId, userId, channelId)
	if err != nil {
		return false
	}

	if permission.HasPermissionRaw(sum, permission.Administrator) {
		return true
	}

	hasPermission := true

	for _, perm := range permissions {
		if !permission.HasPermissionRaw(sum, perm) {
			hasPermission = false
			break
		}
	}

	return hasPermission
}

func HasPermissions(ctx *worker.Context, guildId, userId uint64, permissions ...permission.Permission) bool {
	sum, err := getEffectivePermissions(ctx, guildId, userId)
	if err != nil {
		sentry.Error(err)
		return false
	}

	if permission.HasPermissionRaw(sum, permission.Administrator) {
		return true
	}

	hasPermission := true

	for _, perm := range permissions {
		if !permission.HasPermissionRaw(sum, perm) {
			hasPermission = false
			break
		}
	}

	return hasPermission
}

func getAllPermissionsChannel(ctx *worker.Context, guildId, userId, channelId uint64) []permission.Permission {
	permissions := make([]permission.Permission, 0)

	sum, err := getEffectivePermissionsChannel(ctx, guildId, userId, channelId)
	if err != nil {
		sentry.Error(err)
		return permissions
	}

	for _, perm := range permission.AllPermissions {
		if permission.HasPermissionRaw(sum, perm) {
			permissions = append(permissions, perm)
		}
	}

	return permissions
}

func getAllPermissions(ctx *worker.Context, guildId, userId uint64) []permission.Permission {
	permissions := make([]permission.Permission, 0)

	sum, err := getEffectivePermissions(ctx, guildId, userId)
	if err != nil {
		sentry.Error(err)
		return permissions
	}

	for _, perm := range permission.AllPermissions {
		if permission.HasPermissionRaw(sum, perm) {
			permissions = append(permissions, perm)
		}
	}

	return permissions
}

func getEffectivePermissionsChannel(ctx *worker.Context, guildId, userId, channelId uint64) (uint64, error) {
	permissions, err := getBasePermissions(ctx, guildId)
	if err != nil {
		return 0, err
	}

	permissions, err = getGuildTotalRolePermissions(ctx, guildId, userId, permissions)
	if err != nil {
		return 0, err
	}

	permissions, err = getChannelBasePermissions(ctx, guildId, channelId, permissions)
	if err != nil {
		return 0, err
	}

	permissions, err = getChannelTotalRolePermissions(ctx, guildId, userId, channelId, permissions)
	if err != nil {
		return 0, err
	}

	permissions, err = getChannelMemberPermissions(ctx, userId, channelId, permissions)
	if err != nil {
		return 0, err
	}

	return permissions, nil
}

func getEffectivePermissions(ctx *worker.Context, guildId, userId uint64) (uint64, error) {
	permissions, err := getBasePermissions(ctx, guildId)
	if err != nil {
		return 0, err
	}

	permissions, err = getGuildTotalRolePermissions(ctx, guildId, userId, permissions)
	if err != nil {
		return 0, err
	}

	return permissions, nil
}

func getChannelMemberPermissions(ctx *worker.Context, userId, channelId uint64, initialPermissions uint64) (uint64, error) {
	ch, err := ctx.GetChannel(channelId)
	if err != nil {
		return 0, err
	}

	for _, overwrite := range ch.PermissionOverwrites {
		if overwrite.Type == channel.PermissionTypeMember && overwrite.Id == userId {
			initialPermissions &= ^overwrite.Deny
			initialPermissions |= overwrite.Allow
		}
	}

	return initialPermissions, nil
}

func getChannelTotalRolePermissions(ctx *worker.Context, guildId, userId, channelId uint64, initialPermissions uint64) (uint64, error) {
	member, err := ctx.GetGuildMember(guildId, userId)
	if err != nil {
		return 0, err
	}

	roles, err := ctx.GetGuildRoles(guildId)
	if err != nil {
		return 0, err
	}

	ch, err := ctx.GetChannel(channelId)
	if err != nil {
		return 0, err
	}

	var allow, deny uint64

	for _, memberRole := range member.Roles {
		for _, role := range roles {
			if memberRole == role.Id {
				for _, overwrite := range ch.PermissionOverwrites {
					if overwrite.Type == channel.PermissionTypeRole && overwrite.Id == role.Id {
						allow |= overwrite.Allow
						deny |= overwrite.Deny
						break
					}
				}
			}
		}
	}

	initialPermissions &= ^deny
	initialPermissions |= allow

	return initialPermissions, nil
}

func getChannelBasePermissions(ctx *worker.Context, guildId, channelId uint64, initialPermissions uint64) (uint64, error) {
	roles, err := ctx.GetGuildRoles(guildId)
	if err != nil {
		return 0, err
	}

	var publicRole *guild.Role
	for _, role := range roles {
		if role.Id == guildId {
			publicRole = &role
			break
		}
	}

	if publicRole == nil {
		return 0, errors.New("couldn't find public role")
	}

	ch, err := ctx.GetChannel(channelId)
	if err != nil {
		return 0, err
	}

	for _, overwrite := range ch.PermissionOverwrites {
		if overwrite.Type == channel.PermissionTypeRole && overwrite.Id == publicRole.Id {
			initialPermissions &= ^overwrite.Deny
			initialPermissions |= overwrite.Allow
			break
		}
	}

	return initialPermissions, nil
}

func getGuildTotalRolePermissions(ctx *worker.Context, guildId, userId uint64, initialPermissions uint64) (uint64, error) {
	member, err := ctx.GetGuildMember(guildId, userId)
	if err != nil {
		return 0, err
	}

	roles, err := ctx.GetGuildRoles(guildId)
	if err != nil {
		return 0, err
	}

	for _, memberRole := range member.Roles {
		for _, role := range roles {
			if memberRole == role.Id {
				initialPermissions |= role.Permissions
			}
		}
	}

	return initialPermissions, nil
}

func getBasePermissions(ctx *worker.Context, guildId uint64) (uint64, error) {
	roles, err := ctx.GetGuildRoles(guildId)
	if err != nil {
		return 0, err
	}

	var publicRole *guild.Role
	for _, role := range roles {
		if role.Id == guildId {
			publicRole = &role
			break
		}
	}

	if publicRole == nil {
		return 0, errors.New("couldn't find public role")
	}

	return publicRole.Permissions, nil
}
