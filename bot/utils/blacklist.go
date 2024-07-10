package utils

import (
	"context"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/objects/member"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
)

// Get whether the user is blacklisted at either global or server level
func IsBlacklisted(ctx context.Context, guildId, userId uint64, member member.Member, permLevel permission.PermissionLevel) (bool, error) {
	// Optimise as much as possible, skip errgroup if we can
	if guildId == 0 {
		blacklisted, err := dbclient.Client.GlobalBlacklist.IsBlacklisted(ctx, userId)
		if err != nil {
			return false, err
		}

		return blacklisted, nil
	} else {
		globalBlacklisted := false
		guildBlacklisted := atomic.NewBool(false)

		group, _ := errgroup.WithContext(ctx)
		group.Go(func() error {
			tmp, err := dbclient.Client.Blacklist.IsBlacklisted(ctx, guildId, userId)
			if err != nil {
				return err
			}

			if tmp {
				guildBlacklisted.Store(true)
			}

			return nil
		})

		group.Go(func() error {
			tmp, err := dbclient.Client.GlobalBlacklist.IsBlacklisted(ctx, userId)
			if err != nil {
				return err
			}

			if tmp {
				globalBlacklisted = true
			}

			return nil
		})

		group.Go(func() error {
			tmp, err := dbclient.Client.RoleBlacklist.IsAnyBlacklisted(ctx, guildId, member.Roles)
			if err != nil {
				return err
			}

			if tmp {
				guildBlacklisted.Store(true)
			}

			return nil
		})

		if err := group.Wait(); err != nil {
			return false, err
		}

		if globalBlacklisted {
			return true, nil
		} else if guildBlacklisted.Load() && permLevel < permission.Support { // Have staff override role blacklist
			return true, nil
		} else {
			return false, nil
		}
	}
}
