package logic

import (
	"context"
	"errors"
	"github.com/TicketsBot/common/collections"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest/request"
	"golang.org/x/sync/errgroup"
	"sync"
)

func HasPermissionForTicket(ctx context.Context, worker *worker.Context, ticket database.Ticket, userId uint64) (bool, error) {
	if ticket.UserId == userId {
		return true, nil
	}

	// Check if user is the guild owner
	guild, err := worker.GetGuild(ticket.GuildId)
	if err != nil {
		return false, err
	}

	if guild.OwnerId == userId {
		return true, nil
	}

	// Get member object
	member, err := worker.GetGuildMember(ticket.GuildId, userId)
	if err != nil {
		return false, err
	}

	// Get admin users and roles
	adminUsers, err := dbclient.Client.Permissions.GetAdmins(ctx, ticket.GuildId)
	if err != nil {
		return false, err
	}

	adminRoles, err := dbclient.Client.RolePermissions.GetAdminRoles(ctx, ticket.GuildId)
	if err != nil {
		return false, err
	}

	// Check if user is admin
	if utils.Contains(adminUsers, userId) {
		return true, nil
	}

	// Check if user has admin role
	for _, roleId := range member.Roles {
		if utils.Contains(adminRoles, roleId) {
			return true, nil
		}
	}

	// Check claim
	claimedBy, err := dbclient.Client.TicketClaims.Get(ctx, ticket.GuildId, ticket.Id)
	if err != nil {
		return false, err
	}

	// If the ticket is claimed
	if claimedBy != 0 {
		if claimedBy == userId {
			return true, nil
		}

		// We have already checked admin users
		return false, nil
	}

	if ticket.PanelId == nil {
		return IsInDefaultTeam(ctx, ticket.GuildId, userId, member)
	} else {
		// Get panel for ticket
		panel, err := dbclient.Client.Panel.GetById(ctx, *ticket.PanelId)
		if err != nil {
			return false, err
		}

		if panel.PanelId == 0 {
			return false, errors.New("Panel not found")
		}

		// Check default team, if assigned to panel
		if panel.WithDefaultTeam {
			hasPermission, err := IsInDefaultTeam(ctx, ticket.GuildId, userId, member)
			if err != nil {
				return false, err
			}

			if hasPermission {
				return true, nil
			}
		}

		// Check whether user is part of a team directly
		teamUsers, err := dbclient.Client.SupportTeamMembers.GetAllSupportMembersForPanel(ctx, panel.PanelId)
		if err != nil {
			return false, err
		}

		if utils.Contains(teamUsers, userId) {
			return true, nil
		}

		// Check whether user has any of the roles
		teamRoles, err := dbclient.Client.SupportTeamRoles.GetAllSupportRolesForPanel(ctx, panel.PanelId)
		if err != nil {
			return false, err
		}

		for _, roleId := range member.Roles {
			if utils.Contains(teamRoles, roleId) {
				return true, nil
			}
		}
	}

	return false, nil
}

func IsInDefaultTeam(ctx context.Context, guildId, userId uint64, member member.Member) (bool, error) {
	// Check users
	supportUsers, err := dbclient.Client.Permissions.GetSupport(ctx, guildId)
	if err != nil {
		return false, err
	}

	if utils.Contains(supportUsers, userId) {
		return true, nil
	}

	// Check roles
	supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(ctx, guildId)
	if err != nil {
		return false, err
	}

	for _, roleId := range member.Roles {
		if utils.Contains(supportRoles, roleId) {
			return true, nil
		}
	}

	return false, nil
}

// FilterStaffMembers Ignores ticket opener
func FilterStaffMembers(
	ctx context.Context,
	worker *worker.Context,
	guildId uint64,
	ticket database.Ticket,
	userIds []uint64,
	excludeBots,
	excludeOpener bool,
) ([]uint64, error) {
	var panel *database.Panel
	if ticket.PanelId != nil {
		tmp, err := dbclient.Client.Panel.GetById(ctx, *ticket.PanelId)
		if err != nil {
			return nil, err
		}

		if tmp.PanelId != 0 && tmp.GuildId == guildId {
			panel = &tmp
		}
	}

	// Retrieve permissions data
	// Get admin users and roles
	adminUsers, err := dbclient.Client.Permissions.GetAdmins(ctx, guildId)
	if err != nil {
		return nil, err
	}

	adminRoles, err := dbclient.Client.RolePermissions.GetAdminRoles(ctx, guildId)
	if err != nil {
		return nil, err
	}

	supportUsers, err := dbclient.Client.Permissions.GetSupport(ctx, guildId)
	if err != nil {
		return nil, err
	}

	supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(ctx, guildId)
	if err != nil {
		return nil, err
	}

	var teamUsers, teamRoles []uint64
	if panel != nil {
		// Check whether user is part of a team directly
		teamUsers, err = dbclient.Client.SupportTeamMembers.GetAllSupportMembersForPanel(ctx, panel.PanelId)
		if err != nil {
			return nil, err
		}

		// Check whether user has any of the roles
		teamRoles, err = dbclient.Client.SupportTeamRoles.GetAllSupportRolesForPanel(ctx, panel.PanelId)
		if err != nil {
			return nil, err
		}
	}

	group, _ := errgroup.WithContext(ctx)

	var staffIds []uint64
	var mu sync.Mutex
	for _, userId := range userIds {
		userId := userId

		if excludeOpener && userId == ticket.UserId {
			continue
		}

		group.Go(func() error {
			member, err := worker.GetGuildMember(guildId, userId)
			if err != nil {
				// If the user has left the server, treat them as no longer staff
				if err, ok := err.(request.RestError); ok && err.StatusCode == 404 {
					return nil
				}

				return err
			}

			if excludeBots && member.User.Bot {
				return nil
			}

			if utils.Contains(adminUsers, userId) || utils.Contains(teamUsers, userId) {
				mu.Lock()
				staffIds = append(staffIds, userId)
				mu.Unlock()
				return nil
			}

			// Check default support team
			if panel == nil || panel.WithDefaultTeam {
				if utils.Contains(supportUsers, userId) {
					mu.Lock()
					staffIds = append(staffIds, userId)
					mu.Unlock()
					return nil
				}

				for _, roleId := range supportRoles {
					if utils.Contains(member.Roles, roleId) {
						mu.Lock()
						staffIds = append(staffIds, userId)
						mu.Unlock()
						return nil
					}
				}
			}

			// Check roles
			for _, roleId := range member.Roles {
				if utils.Contains(adminRoles, roleId) || utils.Contains(teamRoles, roleId) {
					mu.Lock()
					staffIds = append(staffIds, userId)
					mu.Unlock()
					return nil
				}
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return staffIds, nil
}

func GetStaffInThread(ctx context.Context, worker *worker.Context, ticket database.Ticket, threadId uint64) ([]uint64, error) {
	// Calculate how many staff members there are
	members, err := worker.ListThreadMembers(threadId) // TODO: Should we try and maintain a cache
	if err != nil {
		return nil, err
	}

	memberIds := make([]uint64, len(members))
	for i, member := range members {
		memberIds[i] = member.UserId
	}

	staffIds, err := FilterStaffMembers(ctx, worker, ticket.GuildId, ticket, memberIds, true, true)
	if err != nil {
		return nil, err
	}

	return staffIds, nil
}

// GetMemberTeams Returns (default_team, team_ids, error)
func GetMemberTeams(ctx context.Context, worker *worker.Context, guildId, userId uint64) (bool, []int, error) {
	member, err := worker.GetGuildMember(guildId, userId)
	if err != nil {
		return false, nil, err
	}

	return GetMemberTeamsWithMember(ctx, guildId, userId, member)
}

func GetMemberTeamsWithMember(ctx context.Context, guildId, userId uint64, member member.Member) (bool, []int, error) {
	// Determine whether the user is part of the default support team
	supportUsers, err := dbclient.Client.Permissions.GetSupport(ctx, guildId)
	if err != nil {
		return false, nil, err
	}

	supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(ctx, guildId)
	if err != nil {
		return false, nil, err
	}

	defaultSupportTeam := utils.Contains(supportUsers, userId) || utils.HasIntersection(supportRoles, member.Roles)

	// Retrieve IDs of additional support teams
	teamIds := collections.NewSet[int]() // Use set to eliminate duplicate entries

	userTeamIds, err := dbclient.Client.SupportTeamMembers.GetAllTeamsForUser(ctx, guildId, userId)
	if err != nil {
		return false, nil, err
	}

	for _, id := range userTeamIds {
		teamIds.Add(id)
	}

	roleTeamIds, err := dbclient.Client.SupportTeamRoles.GetAllTeamsForRoles(ctx, guildId, member.Roles)
	if err != nil {
		return false, nil, err
	}

	for _, id := range roleTeamIds {
		teamIds.Add(id)
	}

	return defaultSupportTeam, teamIds.Collect(), nil
}
