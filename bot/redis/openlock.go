package redis

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"time"
)

type Mutex interface {
	LockContext(ctx context.Context) error
	UnlockContext(ctx context.Context) (bool, error)
}

const TicketOpenLockExpiry = time.Second * 3

var ErrLockExpired = redsync.ErrLockAlreadyExpired

func TakeTicketOpenLock(ctx context.Context, guildId uint64) (Mutex, error) {
	mu := rs.NewMutex(fmt.Sprintf("tickets:openlock:%d", guildId), redsync.WithExpiry(TicketOpenLockExpiry))
	if err := mu.LockContext(ctx); err != nil {
		return nil, err
	}

	return mu, nil
}
