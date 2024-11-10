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
	var builder strings.Builder
	var inLink bool

	builder.Grow(len(s))

	for i, c := range s {
		if c == ' ' {
			inLink = false
		}

		if !inLink {
			if c == 'h' || c == 'H' {
				if len(s) >= i+8 && strings.EqualFold(s[i:i+8], "https://") {
					inLink = true
				} else if len(s) >= i+7 && strings.EqualFold(s[i:i+7], "http://") {
					inLink = true
				}
			}

			if c == '*' || c == '_' || c == '`' || c == '~' || c == '|' || c == '#' {
				builder.WriteRune('\\')
			}
		}

		builder.WriteRune(c)
	}

	return builder.String()
}
