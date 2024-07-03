package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const IntegrationRoleCacheExpiry = time.Minute * 5

var ErrIntegrationRoleNotCached = errors.New("integration role not cached")

func GetIntegrationRole(ctx context.Context, guildId, botId uint64) (uint64, error) {
	key := fmt.Sprintf("integrationrole:%d:%d", guildId, botId)
	res, err := Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, ErrNil) {
			return 0, ErrIntegrationRoleNotCached
		}

		return 0, err
	}

	roleId, err := strconv.ParseUint(res, 10, 64)
	if err != nil {
		return 0, err
	}

	return roleId, nil
}

func SetIntegrationRole(ctx context.Context, guildId, botId, roleId uint64) error {
	key := fmt.Sprintf("integrationrole:%d:%d", guildId, botId)
	return Client.Set(ctx, key, roleId, IntegrationRoleCacheExpiry).Err()
}

func DeleteIntegrationRole(ctx context.Context, guildId, botId uint64) error {
	key := fmt.Sprintf("integrationrole:%d:%d", guildId, botId)
	return Client.Del(ctx, key).Err()
}
