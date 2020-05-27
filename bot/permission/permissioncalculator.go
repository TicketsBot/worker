package permission

import (
	"errors"
	"github.com/TicketsBot/worker"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/permission"
)

func HasPermissionsChannel(worker *worker.Context, guildId, userId, channelId uint64, permissions ...permission.Permission) bool {
	sum, err := GetEffectivePermissionsChannel(worker, guildId, userId, channelId)
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

func HasPermissions(worker *worker.Context, guildId, userId uint64, permissions ...permission.Permission) bool {
	sum, err := GetEffectivePermissions(worker, guildId, userId)
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

func GetAllPermissionsChannel(worker *worker.Context, guildId, userId, channelId uint64) []permission.Permission {
	permissions := make([]permission.Permission, 0)

	sum, err := GetEffectivePermissionsChannel(worker, guildId, userId, channelId)
	if err != nil {
		return permissions
	}

	for _, perm := range permission.AllPermissions {
		if permission.HasPermissionRaw(sum, perm) {
			permissions = append(permissions, perm)
		}
	}

	return permissions
}

func GetAllPermissions(worker *worker.Context, guildId, userId uint64) []permission.Permission {
	permissions := make([]permission.Permission, 0)

	sum, err := GetEffectivePermissions(worker, guildId, userId)
	if err != nil {
		return permissions
	}

	for _, perm := range permission.AllPermissions {
		if permission.HasPermissionRaw(sum, perm) {
			permissions = append(permissions, perm)
		}
	}

	return permissions
}

func GetEffectivePermissionsChannel(worker *worker.Context, guildId, userId, channelId uint64) (int, error) {
	permissions, err := GetBasePermissions(worker, guildId)
	if err != nil {
		return 0, err
	}

	permissions, err = GetGuildTotalRolePermissions(worker, guildId, userId, permissions)
	if err != nil {
		return 0, err
	}

	permissions, err = GetChannelBasePermissions(worker, guildId, channelId, permissions)
	if err != nil {
		return 0, err
	}

	permissions, err = GetChannelTotalRolePermissions(worker, guildId, userId, channelId, permissions)
	if err != nil {
		return 0, err
	}

	permissions, err = GetChannelMemberPermissions(worker, userId, channelId, permissions)
	if err != nil {
		return 0, err
	}

	return permissions, nil
}

func GetEffectivePermissions(worker *worker.Context, guildId, userId uint64) (int, error) {
	permissions, err := GetBasePermissions(worker, guildId)
	if err != nil {
		return 0, err
	}

	permissions, err = GetGuildTotalRolePermissions(worker, guildId, userId, permissions)
	if err != nil {
		return 0, err
	}

	return permissions, nil
}

func GetChannelMemberPermissions(worker *worker.Context, userId, channelId uint64, initialPermissions int) (int, error) {
	ch, err := worker.GetChannel(channelId)
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

func GetChannelTotalRolePermissions(worker *worker.Context, guildId, userId, channelId uint64, initialPermissions int) (int, error) {
	member, err := worker.GetGuildMember(guildId, userId)
	if err != nil {
		return 0, err
	}

	roles, err := worker.GetGuildRoles(guildId)
	if err != nil {
		return 0, err
	}

	ch, err := worker.GetChannel(channelId)
	if err != nil {
		return 0, err
	}

	allow, deny := 0, 0

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

func GetChannelBasePermissions(worker *worker.Context, guildId, channelId uint64, initialPermissions int) (int, error) {
	roles, err := worker.GetGuildRoles(guildId)
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

	ch, err := worker.GetChannel(channelId)
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

func GetGuildTotalRolePermissions(worker *worker.Context, guildId, userId uint64, initialPermissions int) (int, error) {
	member, err := worker.GetGuildMember(guildId, userId)
	if err != nil {
		return 0, err
	}

	roles, err := worker.GetGuildRoles(guildId)
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

func GetBasePermissions(worker *worker.Context, guildId uint64) (int, error) {
	roles, err := worker.GetGuildRoles(guildId)
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

