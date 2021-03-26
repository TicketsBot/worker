package logic

import (
	"context"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
	"golang.org/x/sync/errgroup"
	"sync"
)

func UpdateCommandPermissions(ctx registry.CommandContext, registry registry.Registry) {
	supportUsers, supportRoles, adminUsers, adminRoles, ownerId := getStaff(ctx)

	var globalCommands []interaction.ApplicationCommand
	if ctx.Worker().IsWhitelabel {
		var err error
		globalCommands, err = ctx.Worker().GetGlobalCommands(ctx.Worker().BotId)
		if err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		globalCommands = bot.GlobalCommands
	}

	data := make([]rest.CommandWithPermissionsData, 0)
outer:
	for _, command := range registry {
		properties := command.Properties()

		if properties.PermissionLevel == permcache.Everyone {
			continue
		}

		for _, globalCommand := range globalCommands {
			if command.Properties().Name == globalCommand.Name {
				var permissions []interaction.ApplicationCommandPermissions

				if properties.PermissionLevel == permcache.Admin {
					for _, userId := range adminUsers {
						permissionData := interaction.ApplicationCommandPermissions{
							Id:         userId,
							Type:       interaction.ApplicationCommandPermissionTypeUser,
							Permission: true,
						}

						permissions = append(permissions, permissionData)
					}

					for _, roleId := range adminRoles {
						permissionData := interaction.ApplicationCommandPermissions{
							Id:         roleId,
							Type:       interaction.ApplicationCommandPermissionTypeRole,
							Permission: true,
						}

						permissions = append(permissions, permissionData)
					}
				} else if properties.PermissionLevel == permcache.Support {
					for _, userId := range supportUsers {
						permissionData := interaction.ApplicationCommandPermissions{
							Id:         userId,
							Type:       interaction.ApplicationCommandPermissionTypeUser,
							Permission: true,
						}

						permissions = append(permissions, permissionData)
					}

					for _, roleId := range supportRoles {
						permissionData := interaction.ApplicationCommandPermissions{
							Id:         roleId,
							Type:       interaction.ApplicationCommandPermissionTypeRole,
							Permission: true,
						}

						permissions = append(permissions, permissionData)
					}
				} else { // ???
					continue outer
				}

				permissions = append(permissions, interaction.ApplicationCommandPermissions{
					Id:         ownerId,
					Type:       interaction.ApplicationCommandPermissionTypeUser,
					Permission: true,
				})

				data = append(data, rest.CommandWithPermissionsData{
					Id:          globalCommand.Id,
					Permissions: permissions,
				})
			}
		}
	}

	if _, err := ctx.Worker().EditBulkCommandPermissions(ctx.Worker().BotId, ctx.GuildId(), data); err != nil {
		ctx.HandleError(err)
		return
	}
}

func getStaff(ctx registry.CommandContext) (supportUsers, supportRoles, adminUsers, adminRoles []uint64, ownerId uint64) {
	var mu sync.Mutex

	group, _ := errgroup.WithContext(context.Background())

	group.Go(func() error {
		guild, err := ctx.Guild()
		if err != nil {
			return err
		}

		ownerId = guild.OwnerId
		return nil
	})

	group.Go(func() error {
		users, err := dbclient.Client.Permissions.GetSupport(ctx.GuildId())
		if err != nil {
			return err
		}

		mu.Lock()
		defer mu.Unlock()

		supportUsers = append(supportUsers, users...)
		return nil
	})

	group.Go(func() error {
		roles, err := dbclient.Client.RolePermissions.GetSupportRoles(ctx.GuildId())
		if err != nil {
			return err
		}

		mu.Lock()
		defer mu.Unlock()

		supportRoles = append(supportRoles, roles...)
		return nil
	})

	group.Go(func() error {
		users, err := dbclient.Client.SupportTeamMembers.GetAllSupportMembers(ctx.GuildId())
		if err != nil {
			return err
		}

		mu.Lock()
		defer mu.Unlock()

		supportUsers = append(supportUsers, users...)
		return nil
	})

	group.Go(func() error {
		roles, err := dbclient.Client.SupportTeamRoles.GetAllSupportRoles(ctx.GuildId())
		if err != nil {
			return err
		}

		mu.Lock()
		defer mu.Unlock()

		supportRoles = append(supportRoles, roles...)
		return nil
	})

	group.Go(func() error {
		users, err := dbclient.Client.Permissions.GetAdmins(ctx.GuildId())
		if err != nil {
			return err
		}

		// no lock needed
		adminUsers = append(adminUsers, users...)
		return nil
	})

	group.Go(func() error {
		roles, err := dbclient.Client.RolePermissions.GetAdminRoles(ctx.GuildId())
		if err != nil {
			return err
		}

		// no lock needed
		adminRoles = append(adminRoles, roles...)
		return nil
	})

	if err := group.Wait(); err != nil {
		ctx.HandleError(err)
		return
	}

	return
}
