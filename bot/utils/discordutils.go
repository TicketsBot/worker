package utils

import "time"

const DiscordEpoch uint64 = 1420070400000

func SnowflakeToTime(snowflake uint64) time.Time {
	return time.UnixMilli(int64((snowflake >> 22) + DiscordEpoch))
}
