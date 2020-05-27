package sentry

import (
	"github.com/getsentry/raven-go"
	"github.com/go-errors/errors"
	"os"
	"time"
)

func Connect() {
	if err := raven.SetDSN(os.Getenv("WORKER_SENTRY_DSN")); err != nil {
		Error(err)
		return
	}
}

func ConstructErrorPacket(e *errors.Error) *raven.Packet {
	return ConstructPacket(e, raven.ERROR)
}

func ConstructPacket(e *errors.Error, level raven.Severity) *raven.Packet {
	hostname, err := os.Hostname(); if err != nil {
		hostname = "null"
		Error(err)
	}

	extra := map[string]interface{}{
		"stack": e.ErrorStack(),
	}

	return &raven.Packet{
		Message: e.Error(),
		Extra: extra,
		Project: "tickets-bot",
		Timestamp: raven.Timestamp(time.Now()),
		Level: level,
		ServerName: hostname,
	}
}

