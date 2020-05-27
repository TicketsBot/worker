package sentry

import (
	"github.com/getsentry/raven-go"
	"github.com/go-errors/errors"
	"strconv"
)

type ErrorContext struct {
	Guild       uint64
	User        uint64
	Channel     uint64
	Shard       int
	Command     string
	PremiumTier int
}

func Error(e error) {
	wrapped := errors.New(e)
	raven.Capture(ConstructErrorPacket(wrapped), nil)
}

func LogWithContext(e error, ctx ErrorContext) {
	wrapped := errors.New(e)
	raven.Capture(ConstructPacket(wrapped, raven.INFO), map[string]string{
		"guild":   strconv.FormatUint(ctx.Guild, 10),
		"user":    strconv.FormatUint(ctx.User, 10),
		"channel": strconv.FormatUint(ctx.Channel, 10),
		"shard":   strconv.Itoa(ctx.Shard),
		"command": ctx.Command,
		"premium": strconv.Itoa(ctx.PremiumTier),
	})
}

func LogRestRequest(url string) {
	raven.CaptureMessage(url, nil, nil)
}

func ErrorWithContext(e error, ctx ErrorContext) {
	wrapped := errors.New(e)
	raven.Capture(ConstructErrorPacket(wrapped), map[string]string{
		"guild":   strconv.FormatUint(ctx.Guild, 10),
		"user":    strconv.FormatUint(ctx.User, 10),
		"channel": strconv.FormatUint(ctx.Channel, 10),
		"shard":   strconv.Itoa(ctx.Shard),
		"command": ctx.Command,
		"premium": strconv.Itoa(ctx.PremiumTier),
	})
}
