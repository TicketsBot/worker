package logic

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

// each msg is
const perField = 8

func BuildViewStaffMessage(ctx context.Context, cmd registry.CommandContext, page int) (*embed.Embed, bool) {
	isBlank := true

	self, _ := cmd.Worker().Self()
	embed := embed.NewEmbed().
		SetColor(cmd.GetColour(customisation.Green)).
		SetTitle("Staff").
		SetFooter(fmt.Sprintf("Page %d", page+1), self.AvatarUrl(256))

	// Add field for admin users
	{
		adminUsers, err := dbclient.Client.Permissions.GetAdmins(ctx, cmd.GuildId())
		if err != nil {
			sentry.ErrorWithContext(err, cmd.ToErrorContext())
		}

		lower := perField * page
		upper := perField * (page + 1)

		if lower >= len(adminUsers) {
			embed.AddField("Admin Users", "No admin users", true)
		} else {
			if upper >= len(adminUsers) {
				upper = len(adminUsers)
			}

			var content string
			for i := lower; i < upper; i++ {
				userId := adminUsers[i]
				content += fmt.Sprintf("• <@%d> (`%d`)\n", userId, userId)
			}
			content = strings.TrimSuffix(content, "\n")

			embed.AddField("Admin Users", content, true)
			isBlank = false
		}
	}

	// Add field for admin roles
	{
		adminRoles, err := dbclient.Client.RolePermissions.GetAdminRoles(ctx, cmd.GuildId())
		if err != nil {
			sentry.ErrorWithContext(err, cmd.ToErrorContext())
		}

		lower := perField * page
		upper := perField * (page + 1)

		if lower >= len(adminRoles) {
			embed.AddField("Admin Roles", "No admin roles", true)
		} else {
			if upper >= len(adminRoles) {
				upper = len(adminRoles)
			}

			var content string
			for i := lower; i < upper; i++ {
				roleId := adminRoles[i]
				content += fmt.Sprintf("• <@&%d> (`%d`)\n", roleId, roleId)
			}
			content = strings.TrimSuffix(content, "\n")

			embed.AddField("Admin Roles", content, true)
			isBlank = false
		}
	}

	embed.AddBlankField(false) // Add spacer between admin & support reps

	// Add field for support representatives
	{
		supportUsers, err := dbclient.Client.Permissions.GetSupportOnly(ctx, cmd.GuildId())
		if err != nil {
			sentry.ErrorWithContext(err, cmd.ToErrorContext())
		}

		lower := perField * page
		upper := perField * (page + 1)

		if lower >= len(supportUsers) {
			embed.AddField("Support Representatives", "No support representatives", true)
		} else {
			if upper >= len(supportUsers) {
				upper = len(supportUsers)
			}

			content := "**Warning:** Users in support teams are now deprecated. Please migrate to roles.\n\n"
			for i := lower; i < upper; i++ {
				userId := supportUsers[i]
				content += fmt.Sprintf("• <@%d> (`%d`)\n", userId, userId)
			}
			content = strings.TrimSuffix(content, "\n")

			embed.AddField("Support Representatives", content, true)
			isBlank = false
		}
	}

	// Add field for support roles
	{
		supportRoles, err := dbclient.Client.RolePermissions.GetSupportRolesOnly(ctx, cmd.GuildId())
		if err != nil {
			sentry.ErrorWithContext(err, cmd.ToErrorContext())
		}

		lower := perField * page
		upper := perField * (page + 1)

		if lower >= len(supportRoles) {
			embed.AddField("Support Roles", "No support roles", true)
		} else {
			if upper >= len(supportRoles) {
				upper = len(supportRoles)
			}

			var content string
			for i := lower; i < upper; i++ {
				roleId := supportRoles[i]
				content += fmt.Sprintf("• <@&%d> (`%d`)\n", roleId, roleId)
			}
			content = strings.TrimSuffix(content, "\n")

			embed.AddField("Support Roles", content, true)
			isBlank = false
		}
	}

	return embed, isBlank
}
