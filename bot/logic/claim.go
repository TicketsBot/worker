package logic

import (
	"context"
	"errors"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"golang.org/x/sync/errgroup"
	"sync"
)

func ClaimTicket(worker *worker.Context, ticket database.Ticket, userId uint64) error {
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   ticket.GuildId,
		User:    userId,
	}

	if ticket.ChannelId == nil {
		return errors.New("channel ID is nil")
	}

	// Set to claimed in DB
	if err := dbclient.Client.TicketClaims.Set(ticket.GuildId, ticket.Id, userId); err != nil {
		return err
	}

	// Get claim settings for guild
	claimSettings, err := dbclient.Client.ClaimSettings.Get(ticket.GuildId)
	if err != nil {
		return err
	}

	adminUsers, err := dbclient.Client.Permissions.GetAdmins(ticket.GuildId)
	if err != nil {
		return err
	}

	adminRoles, err := dbclient.Client.RolePermissions.GetAdminRoles(ticket.GuildId)
	if err != nil {
		return err
	}

	var newOverwrites []channel.PermissionOverwrite
	if !claimSettings.SupportCanView {
		newOverwrites = overwritesCantView(userId, worker.BotId, ticket.UserId, ticket.GuildId, adminUsers, adminRoles)
	} else if !claimSettings.SupportCanType {
		// TODO: Teams
		supportUsers, err := dbclient.Client.Permissions.GetSupportOnly(ticket.GuildId)
		if err != nil {
			return err
		}

		supportRoles, err := dbclient.Client.RolePermissions.GetSupportRolesOnly(ticket.GuildId)
		if err != nil {
			return err
		}

		if ticket.PanelId != nil {
			teams, err := dbclient.Client.PanelTeams.GetTeams(*ticket.PanelId)
			if err != nil {
				sentry.ErrorWithContext(err, errorContext)
			} else {
				group, _ := errgroup.WithContext(context.Background())
				mu := sync.Mutex{}

				for _, team := range teams {
					team := team

					// TODO: Joins
					group.Go(func() error {
						members, err := dbclient.Client.SupportTeamMembers.Get(team.Id)
						if err != nil {
							return err
						}

						roles, err := dbclient.Client.SupportTeamRoles.Get(team.Id)
						if err != nil {
							return err
						}

						mu.Lock()
						defer mu.Unlock()
						supportUsers = append(supportUsers, members...)
						supportRoles = append(supportRoles, roles...)

						return nil
					})
				}

				if err := group.Wait(); err != nil {
					sentry.ErrorWithContext(err, errorContext)
				}
			}
		}

		newOverwrites = overwritesCantType(userId, worker.BotId, ticket.UserId, ticket.GuildId, supportUsers, supportRoles, adminUsers, adminRoles)
	}

	// Update channel
	data := rest.ModifyChannelData{
		PermissionOverwrites: newOverwrites,
	}
	if _, err = worker.ModifyChannel(*ticket.ChannelId, data); err != nil {
		return err
	}

	return nil
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
			Allow: permission.BuildPermissions(allowedPermissions...),
			Deny:  0,
		})
	}

	for _, roleId := range adminRoles {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    roleId,
			Type:  channel.PermissionTypeRole,
			Allow: permission.BuildPermissions(allowedPermissions...),
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
			Allow: permission.BuildPermissions(allowedPermissions...),
			Deny:  0,
		})
	}

	for _, roleId := range adminRoles {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    roleId,
			Type:  channel.PermissionTypeRole,
			Allow: permission.BuildPermissions(allowedPermissions...),
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
