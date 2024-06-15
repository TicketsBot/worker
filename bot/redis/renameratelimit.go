package redis

import (
	"context"
	"fmt"
	"time"
)

const (
	renameRatelimitExpiry = time.Minute * 10
	renameRatelimitTokens = 2
)

func TakeRenameRatelimit(ctx context.Context, channelId uint64) (bool, error) {
	key := fmt.Sprintf("tickets:rename_ratelimit:%d", channelId)

	tx := Client.TxPipeline()
	tx.SetNX(ctx, key, "0", renameRatelimitExpiry)
	incr := tx.Incr(ctx, key)

	if _, err := tx.Exec(ctx); err != nil {
		return false, err
	}

	count, err := incr.Result()
	if err != nil {
		return false, err
	}

	return count <= renameRatelimitTokens, nil
}
