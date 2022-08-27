package command

import (
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/redis"
)

func LoadCommandIds(worker *worker.Context, botId uint64) (map[string]uint64, error) {
	// Check cache first
	cached, err := redis.LoadCommandIds(botId)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	// Not cached
	commands, err := worker.GetGlobalCommands(botId) // TODO: Do we store guild commands?
	if err != nil {
		return nil, err
	}

	mapped := make(map[string]uint64)
	for _, command := range commands {
		mapped[command.Name] = command.Id
	}

	if err := redis.StoreCommandIds(botId, mapped); err != nil {
		return nil, err
	}

	return mapped, nil
}
