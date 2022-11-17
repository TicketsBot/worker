package utils

import "time"

func SnowflakeToTime(snowflake uint64) time.Time {
	return time.UnixMilli(int64((snowflake >> 22) + 1420070400000))
}
