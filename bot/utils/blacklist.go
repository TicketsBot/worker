package utils

import (
	"context"
	"github.com/TicketsBot/worker/bot/dbclient"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
)

// Get whether the user is blacklisted at either global or server level
func IsBlacklisted(guildId, userId uint64) (bool, error) {
	// Optimise as much as possible, skip errgroup if we can
	if guildId == 0 {
		blacklisted, err := dbclient.Client.GlobalBlacklist.IsBlacklisted(userId)
		if err != nil {
			return false, err
		}

		return blacklisted, nil
	} else {
		blacklisted := atomic.NewBool(false)

		group, _ := errgroup.WithContext(context.Background())
		group.Go(func() error {
			tmp, err := dbclient.Client.Blacklist.IsBlacklisted(guildId, userId)
			if err != nil {
				return err
			}

			if tmp {
				blacklisted.Store(true)
			}

			return nil
		})

		group.Go(func() error {
			tmp, err := dbclient.Client.GlobalBlacklist.IsBlacklisted(userId)
			if err != nil {
				return err
			}

			if tmp {
				blacklisted.Store(true)
			}

			return nil
		})

		if err := group.Wait(); err != nil {
			return false, err
		}

		return blacklisted.Load(), nil
	}
}
