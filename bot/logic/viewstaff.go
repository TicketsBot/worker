package logic

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

// each msg is
const perField = 16

func BuildViewStaffMessage(guildId uint64, page int, errorContext sentry.ErrorContext) *embed.Embed {
	embed := embed.NewEmbed().
		SetColor(int(utils.Green)).
		SetTitle("Staff")

	// Add field for admin users
	{
		adminUsers, err := dbclient.Client.Permissions.GetAdmins(guildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
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
		}
	}

	// Add field for admin roles
	{
		adminRoles, err := dbclient.Client.RolePermissions.GetAdminRoles(guildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
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
		}
	}

	embed.AddBlankField(false) // Add spacer between admin & support reps

	// Add field for support representatives
	{
		supportUsers, err := dbclient.Client.Permissions.GetSupportOnly(guildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}

		lower := perField * page
		upper := perField * (page + 1)

		if lower >= len(supportUsers) {
			embed.AddField("Support Representatives", "No support representatives", true)
		} else {
			if upper >= len(supportUsers) {
				upper = len(supportUsers)
			}

			var content string
			for i := lower; i < upper; i++ {
				userId := supportUsers[i]
				content += fmt.Sprintf("• <@%d> (`%d`)\n", userId, userId)
			}
			content = strings.TrimSuffix(content, "\n")

			embed.AddField("Support Representatives", content, true)
		}
	}

	// Add field for support roles
	{
		supportRoles, err := dbclient.Client.RolePermissions.GetSupportRolesOnly(guildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
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
		}
	}

	return embed
}
