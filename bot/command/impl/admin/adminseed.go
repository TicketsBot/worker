package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/rest"
)

type AdminSeedCommand struct {
}

func (AdminSeedCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "seed",
		Description:     i18n.HelpAdminSeed,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly:     true,
	}
}

func (c AdminSeedCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminSeedCommand) Execute(ctx registry.CommandContext) {
	var guilds []uint64
	guilds = []uint64{ctx.GuildId()}

	ctx.ReplyRaw(utils.Green, "Admin", fmt.Sprintf("Seeding %d guild(s)", len(guilds)))

	// retrieve all guild members
	var seeded int
	for _, guildId := range guilds {
		moreAvailable := true
		after := uint64(0)

		for moreAvailable {
			// calling this func will cache for us
			members, _ := ctx.Worker().ListGuildMembers(guildId, rest.ListGuildMembersData{
				Limit: 1000,
				After: after,
			})

			if len(members) < 1000 {
				moreAvailable = false
			}

			if len(members) > 0 {
				after = members[len(members) - 1].User.Id
			}
		}

		seeded++

		if seeded % 10 == 0 {
			ctx.ReplyRaw(utils.Green, "Admin", fmt.Sprintf("Seeded %d / %d guilds", seeded, len(guilds)))
		}
	}

	ctx.ReplyRaw(utils.Green, "Admin", "Seeding complete")
}
