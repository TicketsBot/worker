package permissionwrapper

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/permission"
)

func HasPermissionsChannel(ctx *worker.Context, guildId, userId, channelId uint64, permissions ...permission.Permission) bool {
	sum, err := GetEffectivePermissionsChannel(ctx, guildId, userId, channelId)
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
	sum, err := GetEffectivePermissions(ctx, guildId, userId)
	fmt.Println(sum)
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

func GetAllPermissionsChannel(ctx *worker.Context, guildId, userId, channelId uint64) []permission.Permission {
	permissions := make([]permission.Permission, 0)

	sum, err := GetEffectivePermissionsChannel(ctx, guildId, userId, channelId)
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

func GetAllPermissions(ctx *worker.Context, guildId, userId uint64) []permission.Permission {
	permissions := make([]permission.Permission, 0)

	sum, err := GetEffectivePermissions(ctx, guildId, userId)
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

func GetEffectivePermissionsChannel(ctx *worker.Context, guildId, userId, channelId uint64) (uint64, error) {
	permissions, err := GetBasePermissions(ctx, guildId)
	if err != nil {
		return 0, err
	}

	permissions, err = GetGuildTotalRolePermissions(ctx, guildId, userId, permissions)
	if err != nil {
		return 0, err
	}

	permissions, err = GetChannelBasePermissions(ctx, guildId, channelId, permissions)
	if err != nil {
		return 0, err
	}

	permissions, err = GetChannelTotalRolePermissions(ctx, guildId, userId, channelId, permissions)
	if err != nil {
		return 0, err
	}

	permissions, err = GetChannelMemberPermissions(ctx, userId, channelId, permissions)
	if err != nil {
		return 0, err
	}

	return permissions, nil
}

func GetEffectivePermissions(ctx *worker.Context, guildId, userId uint64) (uint64, error) {
	permissions, err := GetBasePermissions(ctx, guildId)
	if err != nil {
		return 0, err
	}

	permissions, err = GetGuildTotalRolePermissions(ctx, guildId, userId, permissions)
	if err != nil {
		return 0, err
	}

	return permissions, nil
}

func GetChannelMemberPermissions(ctx *worker.Context, userId, channelId uint64, initialPermissions uint64) (uint64, error) {
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

func GetChannelTotalRolePermissions(ctx *worker.Context, guildId, userId, channelId uint64, initialPermissions uint64) (uint64, error) {
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

func GetChannelBasePermissions(ctx *worker.Context, guildId, channelId uint64, initialPermissions uint64) (uint64, error) {
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

func GetGuildTotalRolePermissions(ctx *worker.Context, guildId, userId uint64, initialPermissions uint64) (uint64, error) {
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

func GetBasePermissions(ctx *worker.Context, guildId uint64) (uint64, error) {
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
