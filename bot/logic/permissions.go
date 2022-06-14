package logic

import "github.com/rxdn/gdl/permission"

// StandardPermissions Returns the standard permissions that users are given in a ticket
// TODO: Allow servers to choose whether to give attach files
var StandardPermissions = [...]permission.Permission{
	permission.ViewChannel,
	permission.SendMessages,
	permission.AddReactions,
	permission.AttachFiles,
	permission.ReadMessageHistory,
	permission.EmbedLinks,
	permission.UseApplicationCommands,
}
