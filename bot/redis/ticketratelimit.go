package redis

import (
	"fmt"
	"github.com/TicketsBot/common/utils"
	"github.com/go-redis/redis/v8"
	"time"
)

var script = redis.NewScript(`
local current = redis.call("GET", KEYS[1])
local notExists = (not current)

if not current then
	current = 0
else
	current = tonumber(current)
end

local success = 0
if current < tonumber(ARGV[1]) then
	redis.call("INCR", KEYS[1])

	if notExists then
		redis.call("EXPIRE", KEYS[1], ARGV[2])
	end

	success = 1
end

return success
`)

var TicketOpenLimit = 10
var TicketOpenLimitInterval = time.Second * 30

func TakeTicketRateLimitToken(client *redis.Client, guildId uint64) (bool, error) {
	key := fmt.Sprintf("tickets:openratelimit:%d", guildId)

	res, err := script.Run(utils.DefaultContext(), client, []string{key}, TicketOpenLimit, TicketOpenLimitInterval.Seconds()).Result()
	if err != nil {
		return false, err
	}

	i, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("ratelimit token returned %v, not an int64", res)
	}

	return i == 1, nil
}
