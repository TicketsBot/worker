package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

type ViewStaffCommand struct {
}

func (ViewStaffCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "viewstaff",
		Description:     "Lists the staff members and roles",
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
	}
}

func (ViewStaffCommand) Execute(ctx command.CommandContext) {
	embed := embed.NewEmbed().
		SetColor(int(utils.Green)).
		SetTitle("Staff")

	var fieldContent string // temp var

	// Add field for admin users
	adminUsers, err := dbclient.Client.Permissions.GetAdmins(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	for _, adminUserId := range adminUsers {
		fieldContent += fmt.Sprintf("• <@%d> (`%d`)\n", adminUserId, adminUserId)
	}
	fieldContent = strings.TrimSuffix(fieldContent, "\n")
	if fieldContent == "" {
		fieldContent = "No admin users"
	}
	embed.AddField("Admin Users", fieldContent, true)
	fieldContent = ""

	// get existing guild roles
	allRoles, err := ctx.Worker.GetGuildRoles(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Add field for admin roles
	adminRoles, err := dbclient.Client.RolePermissions.GetAdminRoles(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	for _, adminRoleId := range adminRoles {
		for _, guildRole := range allRoles {
			if guildRole.Id == adminRoleId {
				fieldContent += fmt.Sprintf("• %s (`%d`)\n", guildRole.Name, adminRoleId)
			}
		}
	}
	fieldContent = strings.TrimSuffix(fieldContent, "\n")
	if fieldContent == "" {
		fieldContent = "No admin roles"
	}
	embed.AddField("Admin Roles", fieldContent, true)
	fieldContent = ""

	embed.AddBlankField(false) // Add spacer between admin & support reps

	// Add field for support representatives
	supportUsers, err := dbclient.Client.Permissions.GetSupport(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// TODO: Exclude admins
	for _, supportUserId := range supportUsers {
		fieldContent += fmt.Sprintf("• <@%d> (`%d`)\n", supportUserId, supportUserId)
	}
	fieldContent = strings.TrimSuffix(fieldContent, "\n")
	if fieldContent == "" {
		fieldContent = "No support representatives"
	}
	embed.AddField("Support Representatives", fieldContent, true)
	fieldContent = ""

	// Add field for admin roles
	supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// TODO: Exclude admin roles
	for _, supportRoleId := range supportRoles {
		for _, guildRole := range allRoles {
			if guildRole.Id == supportRoleId {
				fieldContent += fmt.Sprintf("• %s (`%d`)\n", guildRole.Name, supportRoleId)
			}
		}
	}
	fieldContent = strings.TrimSuffix(fieldContent, "\n")
	if fieldContent == "" {
		fieldContent = "No support representative roles"
	}
	embed.AddField("Support Roles", fieldContent, true)
	fieldContent = ""

	msg, err := ctx.Worker.CreateMessageEmbed(ctx.ChannelId, embed)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	} else {
		utils.DeleteAfter(utils.SentMessage{Worker: ctx.Worker, Message: &msg}, 60)
	}
}
