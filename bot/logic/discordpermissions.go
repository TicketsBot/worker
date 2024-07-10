package logic

import (
	"context"
	"fmt"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
)

// StandardPermissions Returns the standard permissions that users are given in a ticket
var StandardPermissions = [...]permission.Permission{
	permission.ViewChannel,
	permission.SendMessages,
	permission.AddReactions,
	permission.AttachFiles,
	permission.ReadMessageHistory,
	permission.EmbedLinks,
	permission.UseApplicationCommands,
}

var MinimalPermissions = [...]permission.Permission{
	permission.ViewChannel,
	permission.SendMessages,
	permission.ReadMessageHistory,
	permission.UseApplicationCommands,
}

func BuildUserOverwrite(userId uint64, additionalPermissions database.TicketPermissions) channel.PermissionOverwrite {
	allow := MinimalPermissions[:]
	var deny []permission.Permission

	if additionalPermissions.AttachFiles {
		allow = append(allow, permission.AttachFiles)
	} else {
		deny = append(deny, permission.AttachFiles)
	}

	if additionalPermissions.EmbedLinks {
		allow = append(allow, permission.EmbedLinks)
	} else {
		deny = append(deny, permission.EmbedLinks)
	}

	if additionalPermissions.AddReactions {
		allow = append(allow, permission.AddReactions)
	} else {
		deny = append(deny, permission.AddReactions)
	}

	return channel.PermissionOverwrite{
		Id:    userId,
		Type:  channel.PermissionTypeMember,
		Allow: permission.BuildPermissions(allow...),
		Deny:  permission.BuildPermissions(deny...),
	}
}

func RemoveOnCallRoles(ctx context.Context, cmd registry.CommandContext, userId uint64) error {
	member, err := cmd.Worker().GetGuildMember(cmd.GuildId(), userId)
	if err != nil {
		return err
	}

	metadata, err := dbclient.Client.GuildMetadata.Get(ctx, cmd.GuildId())
	if err != nil {
		return err
	}

	if metadata.OnCallRole != nil && member.HasRole(*metadata.OnCallRole) {
		if err := cmd.Worker().RemoveGuildMemberRole(cmd.GuildId(), userId, *metadata.OnCallRole); err != nil && !isUnknownRoleError(err) {
			return err
		}
	}

	teams, err := dbclient.Client.SupportTeam.Get(ctx, cmd.GuildId())
	if err != nil {
		return err
	}

	for _, team := range teams {
		if team.OnCallRole != nil && member.HasRole(*team.OnCallRole) {
			if err := cmd.Worker().RemoveGuildMemberRole(cmd.GuildId(), userId, *team.OnCallRole); err != nil && !isUnknownRoleError(err) {
				return err
			}
		}
	}

	return nil
}

func RecreateOnCallRole(ctx context.Context, cmd registry.CommandContext, team *database.SupportTeam) error {
	if team == nil {
		metadata, err := dbclient.Client.GuildMetadata.Get(ctx, cmd.GuildId())
		if err != nil {
			return err
		}

		if metadata.OnCallRole == nil {
			return nil
		}

		if err := dbclient.Client.GuildMetadata.SetOnCallRole(ctx, cmd.GuildId(), nil); err != nil {
			return nil
		}

		if err := cmd.Worker().DeleteGuildRole(cmd.GuildId(), *metadata.OnCallRole); err != nil && !isUnknownRoleError(err) {
			return err
		}

		if _, err := CreateOnCallRole(ctx, cmd, nil); err != nil {
			return err
		}

		// TODO: Assign role to on call members
	} else {
		// If there is no on call role, no need to continue
		if team.OnCallRole == nil {
			return nil
		}

		// Delete role
		if err := dbclient.Client.SupportTeam.SetOnCallRole(ctx, team.Id, nil); err != nil {
			return err
		}

		if err := cmd.Worker().DeleteGuildRole(cmd.GuildId(), *team.OnCallRole); err != nil && !isUnknownRoleError(err) {
			return err
		}

		if _, err := CreateOnCallRole(ctx, cmd, team); err != nil {
			return err
		}

		// TODO: Assign role to on call members
	}

	return nil
}

func CreateOnCallRole(ctx context.Context, cmd registry.CommandContext, team *database.SupportTeam) (uint64, error) {
	var roleName string
	if team == nil {
		roleName = "On Call" // TODO: Translate
	} else {
		roleName = utils.StringMax(fmt.Sprintf("On Call - %s", team.Name), 100)
	}

	data := rest.GuildRoleData{
		Name:        roleName,
		Hoist:       utils.Ptr(false),
		Mentionable: utils.Ptr(false),
	}

	role, err := cmd.Worker().CreateGuildRole(cmd.GuildId(), data)
	if err != nil {
		return 0, err
	}

	if team == nil {
		if err := dbclient.Client.GuildMetadata.SetOnCallRole(ctx, cmd.GuildId(), &role.Id); err != nil {
			return 0, err
		}
	} else {
		if err := dbclient.Client.SupportTeam.SetOnCallRole(ctx, team.Id, &role.Id); err != nil {
			return 0, err
		}
	}

	return role.Id, nil
}

func isUnknownRoleError(err error) bool {
	if err, ok := err.(request.RestError); ok && err.ApiError.Message == "Unknown Role" {
		return true
	}

	return false
}
