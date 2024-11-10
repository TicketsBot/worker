package blacklist

import (
	"context"
	"github.com/TicketsBot/worker/bot/dbclient"
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	blacklistedGuilds = make(map[uint64]struct{})
	blacklistedUsers  = make(map[uint64]struct{})
	mu                sync.RWMutex
)

func IsGuildBlacklisted(guildId uint64) bool {
	mu.RLock()
	defer mu.RUnlock()

	_, ok := blacklistedGuilds[guildId]
	return ok
}

func IsUserBlacklisted(userId uint64) bool {
	mu.RLock()
	defer mu.RUnlock()

	_, ok := blacklistedUsers[userId]
	return ok
}

func RefreshCache(ctx context.Context) error {
	guildIds, err := dbclient.Client.ServerBlacklist.ListAll(ctx)
	if err != nil {
		return err
	}

	userIds, err := dbclient.Client.GlobalBlacklist.ListAll(ctx)
	if err != nil {
		return err
	}

	// Build new maps first instead of updating the existing ones to reduce lock time
	guildMap := sliceToMap(guildIds)
	userMap := sliceToMap(userIds)

	mu.Lock()
	defer mu.Unlock()

	blacklistedGuilds = guildMap
	blacklistedUsers = userMap

	return nil
}

func StartCacheRefreshLoop(logger *zap.Logger) {
	logger.Info("Starting blacklist cache refresh loop")

	timer := time.NewTicker(time.Minute * 5)

	for {
		<-timer.C

		if err := RefreshCache(context.Background()); err != nil {
			logger.Error("Failed to refresh blacklist cache", zap.Error(err))
			continue
		}

		logger.Debug("Refreshed blacklist cache successfully")
	}
}

func sliceToMap(slice []uint64) map[uint64]struct{} {
	m := make(map[uint64]struct{})
	for _, v := range slice {
		m[v] = struct{}{}
	}

	return m
}
