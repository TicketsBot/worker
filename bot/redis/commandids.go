package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

func LoadCommandIds(botId uint64) (map[string]uint64, error) {
	data, err := Client.HGetAll(context.Background(), buildCommandIdKey(botId)).Result()
	if err != nil {
		return nil, err
	}

	parsed := make(map[string]uint64)
	for name, idRaw := range data {
		id, err := strconv.ParseUint(idRaw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing id %s for command %s for bot %d", idRaw, name, botId)
		}

		parsed[name] = id
	}

	return parsed, nil
}

func StoreCommandIds(botId uint64, commandIds map[string]uint64) error {
	key := buildCommandIdKey(botId)

	mapped := make(map[string]interface{})
	for name, id := range commandIds {
		mapped[name] = id
	}

	tx := Client.TxPipeline()
	tx.HSet(context.Background(), key, mapped)
	tx.Expire(context.Background(), key, time.Minute*5)

	_, err := tx.Exec(context.Background())
	return err
}

func buildCommandIdKey(botId uint64) string {
	return fmt.Sprintf("commandsids:%d", botId)
}
