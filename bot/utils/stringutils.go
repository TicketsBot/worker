package utils

import "strings"

func StringMax(str string, max int, suffix ...string) string {
	if len(str) > max {
		return str[:max] + strings.Join(suffix, "")
	}

	return str
}