package utils

import (
	"strings"
	"time"
)

const DiscordEpoch uint64 = 1420070400000

func SnowflakeToTime(snowflake uint64) time.Time {
	return time.UnixMilli(int64((snowflake >> 22) + DiscordEpoch))
}

func EscapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "~", "\\~")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "#", "\\#")
	return s
}
