package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"golang.org/x/sync/errgroup"
)

func ClaimTicket(worker *worker.Context, ticket database.Ticket, userId uint64) error {
	if ticket.ChannelId == nil {
		return errors.New("channel ID is nil")
	}

	// Set to claimed in DB
	if err := dbclient.Client.TicketClaims.Set(ticket.GuildId, ticket.Id, userId); err != nil {
		return err
	}

	newOverwrites, err := GenerateClaimedOverwrites(worker, ticket, userId)
	if err != nil {
		return err
	}

	// If newOverwrites = nil, no changes to permissions should be made
	if newOverwrites != nil {
		// Update channel
		data := rest.ModifyChannelData{
			PermissionOverwrites: newOverwrites,
		}
		if _, err = worker.ModifyChannel(*ticket.ChannelId, data); err != nil {
			return err
		}
	}

	return nil
}

// GenerateClaimedOverwrites If support reps can still view and type, returns (nil, nil)
func GenerateClaimedOverwrites(worker *worker.Context, ticket database.Ticket, claimer uint64, otherUsers ...uint64) ([]channel.PermissionOverwrite, error) {
	// Get claim settings for guild
	claimSettings, err := dbclient.Client.ClaimSettings.Get(ticket.GuildId)
	if err != nil {
		return nil, err
	}

	if claimSettings.SupportCanView && claimSettings.SupportCanType {
		return nil, nil
	}

	adminUsers, err := dbclient.Client.Permissions.GetAdmins(ticket.GuildId)
	if err != nil {
		return nil, err
	}

	adminRoles, err := dbclient.Client.RolePermissions.GetAdminRoles(ticket.GuildId)
	if err != nil {
		return nil, err
	}

	// Support can't view the ticket, and therefore can't type either
	if !claimSettings.SupportCanView {
		return overwritesCantView(claimer, worker.BotId, ticket.UserId, ticket.GuildId, adminUsers, adminRoles), nil
	}

	// Support can view the ticket, but can't type
	if !claimSettings.SupportCanType {
		supportUsers, err := dbclient.Client.Permissions.GetSupportOnly(ticket.GuildId)
		if err != nil {
			return nil, err
		}

		supportRoles, err := dbclient.Client.RolePermissions.GetSupportRolesOnly(ticket.GuildId)
		if err != nil {
			return nil, err
		}

		if ticket.PanelId != nil {
			group, _ := errgroup.WithContext(context.Background())

			// Get users for support teams of panel
			group.Go(func() error {
				userIds, err := dbclient.Client.SupportTeamMembers.GetAllSupportMembersForPanel(*ticket.PanelId)
				if err != nil {
					return err
				}

				supportUsers = append(supportUsers, userIds...) // No mutex needed
				return nil
			})

			// Get roles for support teams of panel
			group.Go(func() error {
				roleIds, err := dbclient.Client.SupportTeamRoles.GetAllSupportRolesForPanel(*ticket.PanelId)
				if err != nil {
					return err
				}

				supportRoles = append(supportRoles, roleIds...) // No mutex needed
				return nil
			})

			if err := group.Wait(); err != nil {
				return nil, err
			}
		}

		return overwritesCantType(claimer, worker.BotId, ticket.UserId, ticket.GuildId, supportUsers, supportRoles, adminUsers, adminRoles), nil
	}

	// Unreachable
	return nil, fmt.Errorf("unreachable code reached")
}

// We should build new overwrites from scratch
// TODO: Instead of append(), set indices
func overwritesCantView(claimer, selfId, openerId, guildId uint64, adminUsers, adminRoles []uint64) (overwrites []channel.PermissionOverwrite) {
	overwrites = append(overwrites, channel.PermissionOverwrite{ // @everyone
		Id:    guildId,
		Type:  channel.PermissionTypeRole,
		Allow: 0,
		Deny:  permission.BuildPermissions(permission.ViewChannel),
	})

	for _, userId := range append(adminUsers, claimer, openerId, selfId) {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    userId,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(AllowedPermissions...),
			Deny:  0,
		})
	}

	for _, roleId := range adminRoles {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    roleId,
			Type:  channel.PermissionTypeRole,
			Allow: permission.BuildPermissions(AllowedPermissions...),
			Deny:  0,
		})
	}

	return
}

var readOnlyAllowed = []permission.Permission{permission.ViewChannel, permission.ReadMessageHistory}
var readOnlyDenied = []permission.Permission{permission.SendMessages, permission.AddReactions}

// support & admins are not mutually exclusive due to support teams
func overwritesCantType(claimerId, selfId, openerId, guildId uint64, supportUsers, supportRoles, adminUsers, adminRoles []uint64) (overwrites []channel.PermissionOverwrite) {
	overwrites = append(overwrites, channel.PermissionOverwrite{ // @everyone
		Id:    guildId,
		Type:  channel.PermissionTypeRole,
		Allow: 0,
		Deny:  permission.BuildPermissions(permission.ViewChannel),
	})

	for _, userId := range append(adminUsers, claimerId, selfId, openerId) {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    userId,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(AllowedPermissions...),
			Deny:  0,
		})
	}

	for _, roleId := range adminRoles {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    roleId,
			Type:  channel.PermissionTypeRole,
			Allow: permission.BuildPermissions(AllowedPermissions...),
			Deny:  0,
		})
	}

	for _, userId := range supportUsers {
		// Don't exclude claimer, self or admins
		if userId == claimerId || userId == selfId {
			continue
		}

		for _, adminUserId := range adminUsers {
			if userId == adminUserId {
				continue
			}
		}

		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    userId,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(readOnlyAllowed...),
			Deny:  permission.BuildPermissions(readOnlyDenied...),
		})
	}

	for _, roleId := range supportRoles {
		// Don't exclude claimer, self or admins
		for _, adminRoleId := range adminUsers {
			if roleId == adminRoleId {
				continue
			}
		}

		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    roleId,
			Type:  channel.PermissionTypeRole,
			Allow: permission.BuildPermissions(readOnlyAllowed...),
			Deny:  permission.BuildPermissions(readOnlyDenied...),
		})
	}

	return
}
