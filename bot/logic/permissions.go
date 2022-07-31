package logic

import (
	"github.com/TicketsBot/database"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
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
