package logic

import (
	"errors"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
)

func ClaimTicket(worker *worker.Context, ticket database.Ticket, userId uint64) (err error) {
	if ticket.ChannelId == nil {
		return errors.New("channel ID is nil")
	}

	// Set to claimed in DB
	if err = dbclient.Client.TicketClaims.Set(ticket.GuildId, ticket.Id, userId); err != nil {
		return
	}

	// Get claim settings for guild
	claimSettings, err := dbclient.Client.ClaimSettings.Get(ticket.GuildId)
	if err != nil {
		return
	}

	// Get support users
	supportUsers, err := dbclient.Client.Permissions.GetSupportOnly(ticket.GuildId)
	if err != nil {
		return
	}

	// Get support roles
	supportRoles, err := dbclient.Client.RolePermissions.GetSupportRolesOnly(ticket.GuildId)
	if err != nil {
		return
	}

	// Get existing overwrites
	var overwrites []channel.PermissionOverwrite
	{
		channel, err := worker.GetChannel(*ticket.ChannelId); if err != nil {
		return err
	}

		overwrites = channel.PermissionOverwrites
	}

	// TODO: Just delete from original slice
	var newOverwrites []channel.PermissionOverwrite
	if !claimSettings.SupportCanView {
		newOverwrites = overwritesCantView(overwrites, userId, supportUsers, supportRoles)
	} else if !claimSettings.SupportCanType {
		newOverwrites = overwritesCantType(overwrites, userId, supportUsers, supportRoles)
	}

	// Update channel
	data := rest.ModifyChannelData{
		PermissionOverwrites: newOverwrites,
	}
	if _, err = worker.ModifyChannel(*ticket.ChannelId, data); err != nil {
		return
	}

	return
}

func overwritesCantView(existingOverwrites []channel.PermissionOverwrite, claimer uint64, supportUsers, supportRoles []uint64) (newOverwrites []channel.PermissionOverwrite) {
	var claimerAdded bool

outer:
	for _, overwrite := range existingOverwrites {
		// Remove members
		if overwrite.Type == channel.PermissionTypeMember {
			for _, userId := range supportUsers {
				if overwrite.Id == userId {
					if userId == claimer {
						claimerAdded = true
						break
					} else {
						continue outer
					}
				}
			}

			newOverwrites = append(newOverwrites, overwrite)
		} else if overwrite.Type == channel.PermissionTypeRole { // Remove roles
			for _, roleId := range supportRoles {
				if overwrite.Id == roleId {
					continue outer
				}
			}

			newOverwrites = append(newOverwrites, overwrite)
		}
	}

	if !claimerAdded {
		newOverwrites = append(newOverwrites, channel.PermissionOverwrite{
			Id:    claimer,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
			Deny:  0,
		})
	}

	return
}

func overwritesCantType(existingOverwrites []channel.PermissionOverwrite, claimer uint64, supportUsers, supportRoles []uint64) (newOverwrites []channel.PermissionOverwrite) {
	for _, overwrite := range existingOverwrites {
		// Update members
		if overwrite.Type == channel.PermissionTypeMember {
			for _, userId := range supportUsers {
				if overwrite.Id == userId && overwrite.Id != claimer {
					overwrite.Allow = permission.BuildPermissions(permission.ViewChannel, permission.ReadMessageHistory)
					overwrite.Deny = permission.BuildPermissions(permission.AddReactions, permission.SendMessages)
					break
				}
			}

			newOverwrites = append(newOverwrites, overwrite)
		} else if overwrite.Type == channel.PermissionTypeRole { // Update roles
			for _, roleId := range supportRoles {
				if overwrite.Id == roleId {
					overwrite.Allow = permission.BuildPermissions(permission.ViewChannel, permission.ReadMessageHistory)
					overwrite.Deny = permission.BuildPermissions(permission.AddReactions, permission.SendMessages)
					break
				}
			}

			newOverwrites = append(newOverwrites, overwrite)
		}
	}

	// Add claimer
	newOverwrites = append(newOverwrites, channel.PermissionOverwrite{
		Id:    claimer,
		Type:  channel.PermissionTypeMember,
		Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
		Deny:  0,
	})

	return
}

