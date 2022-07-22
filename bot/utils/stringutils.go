package utils

import (
	"math/rand"
	"strings"
)

func StringMax(str string, max int, suffix ...string) string {
	if len(str) > max {
		return str[:max] + strings.Join(suffix, "")
	}

	return str
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
