package utils

import (
	"golang.org/x/exp/constraints"
	"math/rand"
	"strings"
)

type Number interface {
	constraints.Integer | constraints.Float
}

func StringMax(str string, max int, suffix ...string) string {
	if len(str) > max {
		return str[:max] + strings.Join(suffix, "")
	}

	return str
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	} else {
		return b
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
