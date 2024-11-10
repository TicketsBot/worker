package utils

import (
	"context"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/blacklist"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/objects/member"
	"golang.org/x/sync/errgroup"
)

// Get whether the user is blacklisted at either global or server level
func IsBlacklisted(ctx context.Context, guildId, userId uint64, member member.Member, permLevel permission.PermissionLevel) (bool, error) {
	if blacklist.IsUserBlacklisted(userId) {
		return true, nil
	}

	var userBlacklisted, roleBlacklisted bool

	group, _ := errgroup.WithContext(ctx)
	group.Go(func() error {
		tmp, err := dbclient.Client.Blacklist.IsBlacklisted(ctx, guildId, userId)
		if err != nil {
			return err
		}

		if tmp {
			userBlacklisted = true
		}

		return nil
	})

	group.Go(func() error {
		tmp, err := dbclient.Client.RoleBlacklist.IsAnyBlacklisted(ctx, guildId, member.Roles)
		if err != nil {
			return err
		}

		if tmp {
			roleBlacklisted = true
		}

		return nil
	})

	if err := group.Wait(); err != nil {
		return false, err
	}

	// Have staff override role blacklist
	return permLevel < permission.Support && (userBlacklisted || roleBlacklisted), nil
}
